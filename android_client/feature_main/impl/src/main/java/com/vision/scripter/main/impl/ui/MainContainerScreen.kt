package com.vision.scripter.main.impl.ui

import androidx.compose.foundation.layout.WindowInsets
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Gesture
import androidx.compose.material.icons.filled.Home
import androidx.compose.material.icons.filled.Image
import androidx.compose.material3.Icon
import androidx.compose.material3.NavigationBar
import androidx.compose.material3.NavigationBarItem
import androidx.compose.material3.Scaffold
import androidx.compose.material3.SnackbarHostState
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableIntStateOf
import androidx.compose.runtime.saveable.rememberSaveable
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.res.stringResource
import com.vision.scripter.library.impl.state.LibraryKind
import com.vision.scripter.library.impl.state.LibraryUiStateHolder
import com.vision.scripter.library.impl.ui.LibraryScreen
import com.vision.scripter.main.impl.R
import com.vision.scripter.main.impl.ui.items.MainTopBar
import com.vision.scripter.ui.ProvideSnackbarHost

private const val HomeTab = 0
private const val ImagesTab = 1
private const val ActionsTab = 2

@Composable
internal fun MainContainerScreen(
    mainUiStateHolder: MainUiStateHolder,
    libraryUiStateHolder: LibraryUiStateHolder,
    snackbarHostState: SnackbarHostState,
) {
    var selectedTab by rememberSaveable { mutableIntStateOf(HomeTab) }

    Scaffold(
        modifier = Modifier.fillMaxSize(),
        contentWindowInsets = WindowInsets(0, 0, 0, 0),
        topBar = {
            MainTopBar(onSettingsClick = {})
        },
        snackbarHost = { ProvideSnackbarHost(snackbarHostState) },
        bottomBar = {
            NavigationBar {
                NavigationBarItem(
                    selected = selectedTab == HomeTab,
                    onClick = { selectedTab = HomeTab },
                    icon = {
                        Icon(
                            imageVector = Icons.Default.Home,
                            contentDescription = null,
                        )
                    },
                    label = { Text(text = stringResource(R.string.home)) },
                )
                NavigationBarItem(
                    selected = selectedTab == ImagesTab,
                    onClick = { selectedTab = ImagesTab },
                    icon = {
                        Icon(
                            imageVector = Icons.Default.Image,
                            contentDescription = null,
                        )
                    },
                    label = { Text(text = stringResource(R.string.images)) },
                )
                NavigationBarItem(
                    selected = selectedTab == ActionsTab,
                    onClick = { selectedTab = ActionsTab },
                    icon = {
                        Icon(
                            imageVector = Icons.Default.Gesture,
                            contentDescription = null,
                        )
                    },
                    label = { Text(text = stringResource(R.string.custom_events)) },
                )
            }
        }
    ) { paddingValues ->
        when (selectedTab) {
            ImagesTab -> LibraryScreen(
                uiStateHolder = libraryUiStateHolder,
                kind = LibraryKind.IMAGES,
                paddingValues = paddingValues,
            )

            ActionsTab -> LibraryScreen(
                uiStateHolder = libraryUiStateHolder,
                kind = LibraryKind.ACTIONS,
                paddingValues = paddingValues,
            )

            else -> MainUiScreen(
                uiStateHolder = mainUiStateHolder,
                paddingValues = paddingValues,
            )
        }
    }
}
