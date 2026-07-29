package com.vision.scripter.streaming.impl.screen.main.ui

import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.Scaffold
import androidx.compose.material3.SnackbarHostState
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.unit.dp
import androidx.hilt.lifecycle.viewmodel.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.navigation.NavController
import com.vision.scripter.streaming.impl.R
import com.vision.scripter.streaming.impl.blocks.menu.commandobservers.MenuSharedCommandObserver
import com.vision.scripter.streaming.impl.blocks.menu.state.MenuViewModel
import com.vision.scripter.streaming.impl.blocks.menu.ui.MenuBlock
import com.vision.scripter.streaming.impl.blocks.video.commandobservers.VideoSharedCommandObserver
import com.vision.scripter.streaming.impl.blocks.video.state.VideoViewModel
import com.vision.scripter.streaming.impl.blocks.video.ui.VideoBlock
import com.vision.scripter.ui.CustomButton
import com.vision.scripter.ui.ProvideSnackbarHost

@Composable
internal fun StreamingScreen(
    serial: String,
    uiStateHolder: StreamingUiStateHolder,
    snackbarHostState: SnackbarHostState,
    navController: NavController,
) {
    val state = uiStateHolder.uiStateFlow.collectAsStateWithLifecycle().value
    val menuUiStateHolder = hiltViewModel<MenuViewModel>()
    val videoUiStateHolder = hiltViewModel<VideoViewModel>()
    LaunchedEffect(serial) {
        videoUiStateHolder.init(serial)
    }

    VideoSharedCommandObserver(
        videoUiStateHolder = videoUiStateHolder,
        streamingUiStateHolder = uiStateHolder,
    )

    MenuSharedCommandObserver(
        menuUiStateHolder = menuUiStateHolder,
        streamingUiStateHolder = uiStateHolder,
    )

    Box(modifier = Modifier.fillMaxSize()) {
        VideoBlock(
            modifier = Modifier.fillMaxSize(),
            uiStateHolder = videoUiStateHolder,
        )

        MenuBlock(
            modifier = Modifier
                .align(Alignment.TopEnd)
                .padding(top = 128.dp),
            navController = navController,
            uiStateHolder = menuUiStateHolder,
        )
    }

    if (!state.isError && !state.isLoading) return
    Scaffold(
        modifier = Modifier.fillMaxSize(),
        snackbarHost = { ProvideSnackbarHost(snackbarHostState) },
        topBar = {},
    ) { paddingValues ->
        Box(
            modifier = Modifier
                .fillMaxSize()
                .padding(paddingValues),
        ) {
            if (state.isError) {
                CustomButton(
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(horizontal = 16.dp)
                        .align(Alignment.Center),
                    text = stringResource(R.string.retry),
                    onClick = uiStateHolder::onRefresh,
                )
                return@Scaffold
            }
            CircularProgressIndicator(modifier = Modifier.align(Alignment.Center))
        }
    }
}