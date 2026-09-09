package com.vision.scripter.main.ui

import androidx.compose.foundation.layout.WindowInsets
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Folder
import androidx.compose.material.icons.filled.PhoneAndroid
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
import com.vision.scripter.devices.ui.DevicesScreen
import com.vision.scripter.devices.ui.DevicesUiStateHolder
import com.vision.scripter.library.state.LibraryUiStateHolder
import com.vision.scripter.library.ui.LibraryScreen
import com.vision.scripter.main.impl.R
import com.vision.scripter.main.ui.items.MainTopBar
import com.vision.scripter.ui.ProvideSnackbarHost

private const val DevicesTab = 0
private const val LibraryTab = 1

@Composable
internal fun MainContainerScreen(
    devicesUiStateHolder: DevicesUiStateHolder,
    libraryUiStateHolder: LibraryUiStateHolder,
    snackbarHostState: SnackbarHostState,
) {
    var selectedTab by rememberSaveable { mutableIntStateOf(DevicesTab) }

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
                    selected = selectedTab == DevicesTab,
                    onClick = { selectedTab = DevicesTab },
                    icon = {
                        Icon(
                            imageVector = Icons.Default.PhoneAndroid,
                            contentDescription = null,
                        )
                    },
                    label = { Text(text = stringResource(R.string.devices)) },
                )
                NavigationBarItem(
                    selected = selectedTab == LibraryTab,
                    onClick = { selectedTab = LibraryTab },
                    icon = {
                        Icon(
                            imageVector = Icons.Default.Folder,
                            contentDescription = null,
                        )
                    },
                    label = { Text(text = stringResource(R.string.library)) },
                )
            }
        }
    ) { paddingValues ->
        when (selectedTab) {
            LibraryTab -> LibraryScreen(
                uiStateHolder = libraryUiStateHolder,
                paddingValues = paddingValues,
            )

            else -> DevicesScreen(
                uiStateHolder = devicesUiStateHolder,
                paddingValues = paddingValues,
            )
        }
    }
}
