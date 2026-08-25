package models

import (
	"android_vision_scripter/pkg/core/file"
	"bytes"
	"image"
	"slices"
	"strings"
)

const (
	GridCols           = 18
	GridRows           = 32
	GridMatchThreshold = 0.8
)

type Node struct {
	Name          string         `json:"name"`
	Landmarks     []Landmark     `json:"landmarks"`
	Edges         []Edge         `json:"edges"`
	OccupancyGrid *OccupancyGrid `json:"occupancy_grid,omitempty"`
}

type Edge struct {
	Action    Step     `json:"action"`
	NextNodes []string `json:"next_nodes"`
}

type OccupancyGrid struct {
	Cols  int      `json:"cols"`
	Rows  int      `json:"rows"`
	Cells []string `json:"cells"`
}

func (n *Node) Valid() bool {
	if n == nil || !file.ValidName(n.Name) {
		return false
	}
	for _, landmark := range n.Landmarks {
		if !landmark.Valid() {
			return false
		}
	}
	for _, edge := range n.Edges {
		if !edge.Valid() {
			return false
		}
	}
	if n.OccupancyGrid != nil && !n.OccupancyGrid.Valid() {
		return false
	}
	return true
}

func (n *Node) Merge(update *Node) {
	if update == nil {
		return
	}
	for _, landmark := range update.Landmarks {
		if n.hasLandmark(&landmark) {
			continue
		}
		n.Landmarks = append(n.Landmarks, landmark)
	}
	for _, edge := range update.Edges {
		existing := n.findEdge(&edge.Action)
		if existing == nil {
			n.Edges = append(n.Edges, edge)
			continue
		}
		existing.NextNodes = mergeNames(existing.NextNodes, edge.NextNodes)
	}
	if !update.OccupancyGrid.IsEmpty() {
		n.OccupancyGrid = update.OccupancyGrid
	}
}

func (n *Node) hasLandmark(landmark *Landmark) bool {
	for _, existing := range n.Landmarks {
		if existing.Type == landmark.Type &&
			existing.Value == landmark.Value &&
			existing.Locale == landmark.Locale {
			return true
		}
	}
	return false
}

func (n *Node) findEdge(action *Step) *Edge {
	for index := range n.Edges {
		if n.Edges[index].Action.SameAction(action) {
			return &n.Edges[index]
		}
	}
	return nil
}

func mergeNames(existing []string, update []string) []string {
	for _, name := range update {
		if slices.Contains(existing, name) {
			continue
		}
		existing = append(existing, name)
	}
	return existing
}

func (e *Edge) Valid() bool {
	if e == nil || !e.Action.Valid() || e.Action.IsCheckEvent() {
		return false
	}
	for _, next := range e.NextNodes {
		if !file.ValidName(next) {
			return false
		}
	}
	return true
}

func (g *OccupancyGrid) Valid() bool {
	if g == nil || g.Cols <= 0 || g.Rows <= 0 || len(g.Cells) != g.Rows {
		return false
	}
	for _, row := range g.Cells {
		if len(row) != g.Cols {
			return false
		}
		for _, cell := range row {
			if cell != '0' && cell != '1' {
				return false
			}
		}
	}
	return true
}

func (g *OccupancyGrid) IsEmpty() bool {
	if g == nil || len(g.Cells) == 0 {
		return true
	}
	for _, row := range g.Cells {
		if strings.ContainsRune(row, '1') {
			return false
		}
	}
	return true
}

func (g *OccupancyGrid) Similarity(other *OccupancyGrid) float64 {
	if !g.Valid() || !other.Valid() || g.Cols != other.Cols || g.Rows != other.Rows {
		return 0
	}
	var intersection int
	var union int
	for rowIndex, row := range g.Cells {
		otherRow := other.Cells[rowIndex]
		for colIndex := range row {
			occupied := row[colIndex] == '1'
			otherOccupied := otherRow[colIndex] == '1'
			if occupied && otherOccupied {
				intersection++
			}
			if occupied || otherOccupied {
				union++
			}
		}
	}
	if union == 0 {
		return 0
	}
	return float64(intersection) / float64(union)
}

func (g *OccupancyGrid) Matches(other *OccupancyGrid) bool {
	return g.Similarity(other) >= GridMatchThreshold
}

func GridFromRects(rects []image.Rectangle, width int, height int) *OccupancyGrid {
	rows := make([][]byte, GridRows)
	for index := range rows {
		rows[index] = bytes.Repeat([]byte{'0'}, GridCols)
	}

	if width > 0 && height > 0 {
		for _, rect := range rects {
			startCol := gridClamp(rect.Min.X*GridCols/width, GridCols-1)
			endCol := gridClamp((rect.Max.X-1)*GridCols/width, GridCols-1)
			startRow := gridClamp(rect.Min.Y*GridRows/height, GridRows-1)
			endRow := gridClamp((rect.Max.Y-1)*GridRows/height, GridRows-1)
			for row := startRow; row <= endRow; row++ {
				for col := startCol; col <= endCol; col++ {
					rows[row][col] = '1'
				}
			}
		}
	}

	cells := make([]string, GridRows)
	for index, row := range rows {
		cells[index] = string(row)
	}
	return &OccupancyGrid{Cols: GridCols, Rows: GridRows, Cells: cells}
}

func gridClamp(value int, high int) int {
	if value < 0 {
		return 0
	}
	if value > high {
		return high
	}
	return value
}
