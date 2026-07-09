package com.vision.scripter.main.impl.ui

import androidx.compose.foundation.layout.WindowInsets
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.List
import androidx.compose.material.icons.filled.Home
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
import com.vision.scripter.main.impl.R
import com.vision.scripter.main.impl.ui.items.MainTopBar
import com.vision.scripter.scripts.impl.state.ScriptsUiStateHolder
import com.vision.scripter.scripts.impl.ui.ScriptsScreen
import com.vision.scripter.ui.ProvideSnackbarHost

private const val HomeTab = 0
private const val ScriptsTab = 1

@Composable
internal fun MainContainerScreen(
    mainUiStateHolder: MainUiStateHolder,
    scriptsUiStateHolder: ScriptsUiStateHolder,
    snackbarHostState: SnackbarHostState,
) {
    var selectedTab by rememberSaveable { mutableIntStateOf(HomeTab) }

    Scaffold(
        modifier = Modifier.fillMaxSize(),
        contentWindowInsets = WindowInsets(0, 0, 0, 0),
        topBar = { MainTopBar(onSettingsClick = {}) },
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
                    selected = selectedTab == ScriptsTab,
                    onClick = { selectedTab = ScriptsTab },
                    icon = {
                        Icon(
                            imageVector = Icons.AutoMirrored.Filled.List,
                            contentDescription = null,
                        )
                    },
                    label = { Text(text = stringResource(R.string.scripts)) },
                )
            }
        }
    ) { paddingValues ->
        when (selectedTab) {
            ScriptsTab -> ScriptsScreen(
                uiStateHolder = scriptsUiStateHolder,
                paddingValues = paddingValues,
            )

            else -> MainUiScreen(
                uiStateHolder = mainUiStateHolder,
                paddingValues = paddingValues,
            )
        }
    }
}
