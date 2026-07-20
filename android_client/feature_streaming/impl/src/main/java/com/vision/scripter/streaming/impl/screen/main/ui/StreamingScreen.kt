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
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.vision.scripter.streaming.impl.R
import com.vision.scripter.streaming.impl.blocks.menu.ui.MenuBlock
import com.vision.scripter.streaming.impl.blocks.video.ui.VideoBlock
import com.vision.scripter.ui.CustomButton
import com.vision.scripter.ui.ProvideSnackbarHost

@Composable
internal fun StreamingScreen(
    serial: String,
    uiStateHolder: StreamingUiStateHolder,
    snackbarHostState: SnackbarHostState,
) {
    val state = uiStateHolder.uiStateFlow.collectAsStateWithLifecycle(
        initialValue = StreamingUiState(),
    ).value ?: StreamingUiState()

    LaunchedEffect(Unit) {
        uiStateHolder.initArgs(serial = serial)
        uiStateHolder.onLoadData(onStart = true)
    }

    if (!state.isLoading) {
        Box(modifier = Modifier.fillMaxSize()) {
            VideoBlock(
                modifier = Modifier.fillMaxSize(),
                sharedStateHolder = uiStateHolder,
            )
            MenuBlock(
                modifier = Modifier
                    .align(Alignment.TopEnd)
                    .padding(top = 128.dp),
                sharedStateHolder = uiStateHolder,
            )
        }
        return
    }

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
                    onClick = {
                        uiStateHolder.onLoadData(false)
                    }
                )
                return@Scaffold
            }
            CircularProgressIndicator(modifier = Modifier.align(Alignment.Center))
        }
    }
}