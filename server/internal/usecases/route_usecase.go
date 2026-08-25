package usecases

import (
	"android_vision_scripter/pkg/models"
	"errors"
	"fmt"
	"strings"
	"time"

	"gocv.io/x/gocv"
)

type RouteUseCase interface {
	FollowRoute(serial string, from string, to string, basePort int) error
}

type routeHop struct {
	action       models.Step
	expected     string
	alternatives []string
}

const (
	maxRouteReplans          = 5
	alternativeVerifyTimeout = 2 * time.Second
)

func (i *interactorImpl) FollowRoute(
	serial string,
	from string,
	to string,
	basePort int,
) error {
	serial = strings.TrimSpace(serial)
	if serial == "" {
		return errors.New(SerialIsEmptyError)
	}
	if _, err := i.GetMapNode(from); err != nil {
		return err
	}
	if _, err := i.GetMapNode(to); err != nil {
		return err
	}
	if _, err := i.planRoute(from, to); err != nil {
		return err
	}

	if _, ok := i.sessionsCache.Get(serial); !ok {
		started := i.StartSession(serial, basePort)
		if !started {
			return fmt.Errorf("couldn't start scrcpy server for %s", serial)
		}
		go i.scrcpy.ReadVideoStream(serial, nil)
	}

	added := i.addRouteToQueue(serial, &models.Route{From: from, To: to})
	if !added {
		return fmt.Errorf("no active session for %s", serial)
	}
	return nil
}

func (i *interactorImpl) planRoute(from string, to string) ([]routeHop, error) {
	if from == to {
		return []routeHop{}, nil
	}

	type parentLink struct {
		node string
		hop  routeHop
	}
	visited := map[string]bool{from: true}
	parents := map[string]parentLink{}
	queue := []string{from}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		node, err := i.GetMapNode(current)
		if err != nil {
			continue
		}

		for _, edge := range node.Edges {
			for _, next := range edge.NextNodes {
				if visited[next] {
					continue
				}
				visited[next] = true
				parents[next] = parentLink{
					node: current,
					hop: routeHop{
						action:       edge.Action,
						expected:     next,
						alternatives: otherNames(edge.NextNodes, next),
					},
				}
				if next == to {
					hops := []routeHop{}
					for walk := to; walk != from; walk = parents[walk].node {
						hops = append([]routeHop{parents[walk].hop}, hops...)
					}
					return hops, nil
				}
				queue = append(queue, next)
			}
		}
	}
	return nil, fmt.Errorf("no route from %s to %s", from, to)
}

func otherNames(names []string, exclude string) []string {
	others := []string{}
	for _, name := range names {
		if name == exclude {
			continue
		}
		others = append(others, name)
	}
	return others
}

func (i *interactorImpl) executeRoute(serial string, route *models.Route) error {
	i.setRouteStatus(serial, route, "verifying "+route.From)
	confirmed, err := i.verifyNode(serial, route.From, time.Duration(models.DefaultTimeout)*time.Second, true)
	if err != nil {
		return err
	}
	if !confirmed {
		return i.lostError(route, route.From)
	}

	current := route.From
	replans := 0
	for current != route.To {
		hops, err := i.planRoute(current, route.To)
		if err != nil {
			return err
		}

		for _, hop := range hops {
			i.setRouteStatus(serial, route, hop.action.ToString())
			step := hop.action
			if err := i.executeStep(serial, &step); err != nil {
				return err
			}

			i.setRouteStatus(serial, route, "verifying "+hop.expected)
			confirmed, err := i.verifyNode(serial, hop.expected, time.Duration(models.DefaultTimeout)*time.Second, true)
			if err != nil {
				return err
			}
			if confirmed {
				current = hop.expected
				continue
			}

			landed := i.checkAlternatives(serial, hop.alternatives)
			if landed == "" || landed == current {
				return i.lostError(route, hop.expected)
			}
			replans++
			if replans > maxRouteReplans {
				return i.lostError(route, hop.expected)
			}
			current = landed
			break
		}
	}
	return nil
}

func (i *interactorImpl) verifyNode(
	serial string,
	name string,
	timeout time.Duration,
	trustEmpty bool,
) (bool, error) {
	node, err := i.GetMapNode(name)
	if err != nil {
		return false, err
	}

	hasLandmarks := len(node.Landmarks) > 0
	hasGrid := !node.OccupancyGrid.IsEmpty()
	if !hasLandmarks && !hasGrid {
		return trustEmpty, nil
	}

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		mat, err := i.scrcpy.GetMatFromLastFrame(serial, true)
		if err != nil || mat == nil {
			sleepUntilNextSecond()
			continue
		}

		if hasLandmarks && i.landmarksOnFrame(serial, mat, node.Landmarks) {
			if !hasGrid {
				i.captureGrid(node, mat)
			}
			mat.Close()
			return true, nil
		}

		if hasGrid && i.gridMatchesFrame(node.OccupancyGrid, mat) {
			mat.Close()
			return true, nil
		}

		mat.Close()
		sleepUntilNextSecond()
	}
	return false, nil
}

func (i *interactorImpl) landmarksOnFrame(
	serial string,
	mat *gocv.Mat,
	landmarks []models.Landmark,
) bool {
	for _, landmark := range landmarks {
		candidates, err := i.findLandmarkCandidates(serial, mat, &landmark)
		if err != nil {
			i.logger.Error(err.Error())
			return false
		}
		if len(candidates) == 0 {
			return false
		}
	}
	return true
}

func (i *interactorImpl) captureGrid(node *models.Node, mat *gocv.Mat) {
	rects, err := i.cv.FindStructuralRectangles(mat)
	if err != nil {
		i.logger.Error(err.Error())
		return
	}

	grid := models.GridFromRects(rects, mat.Cols(), mat.Rows())
	if grid.IsEmpty() {
		return
	}

	node.OccupancyGrid = grid
	err = i.writeMapNode(node)
	if err != nil {
		i.logger.Error(err.Error())
	}
}

func (i *interactorImpl) gridMatchesFrame(
	grid *models.OccupancyGrid,
	mat *gocv.Mat,
) bool {
	rects, err := i.cv.FindStructuralRectangles(mat)
	if err != nil {
		return false
	}
	frameGrid := models.GridFromRects(rects, mat.Cols(), mat.Rows())
	return grid.Matches(frameGrid)
}

func (i *interactorImpl) checkAlternatives(serial string, alternatives []string) string {
	for _, name := range alternatives {
		confirmed, err := i.verifyNode(serial, name, alternativeVerifyTimeout, false)
		if err != nil {
			continue
		}
		if confirmed {
			return name
		}
	}
	return ""
}

func (i *interactorImpl) setRouteStatus(
	serial string,
	route *models.Route,
	phase string,
) {
	status := fmt.Sprintf(models.StatusRunningRoute, route.From, route.To, phase)
	i.sessionsCache.Update(serial, func(session models.Session) models.Session {
		session.Status = status
		return session
	})
}

func (i *interactorImpl) lostError(route *models.Route, node string) error {
	return fmt.Errorf(models.StatusLostRoute, route.From, route.To, node)
}
