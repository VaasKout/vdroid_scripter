package com.vision.scripter.scripts.impl.ui

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.Scaffold
import androidx.compose.material3.SnackbarHostState
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.remember
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.tooling.preview.Preview
import androidx.compose.ui.unit.dp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.vision.scripter.scripts.impl.R
import com.vision.scripter.scripts.impl.state.ScriptsUiCommand
import com.vision.scripter.scripts.impl.state.ScriptsUiState
import com.vision.scripter.scripts.impl.state.ScriptsUiStateHolder
import com.vision.scripter.ui.CustomPullToRefresh
import com.vision.scripter.ui.DeleteDialog
import com.vision.scripter.ui.ProvideSnackbarHost

@Composable
internal fun ScriptsScreen(
    uiStateHolder: ScriptsUiStateHolder,
    snackbarHostState: SnackbarHostState,
) {
    val state = uiStateHolder.uiStateFlow.collectAsStateWithLifecycle(
        initialValue = ScriptsUiState(),
    ).value

    LaunchedEffect(Unit) {
        uiStateHolder.onLoadData(onStart = true)
    }

    Scaffold(
        modifier = Modifier.fillMaxSize(),
        snackbarHost = { ProvideSnackbarHost(snackbarHostState) },
        topBar = {
            ScriptsTopBar(onBackClick = {
                uiStateHolder.uiCommandsFlow.tryEmit(ScriptsUiCommand.NavigateBack)
            })
        }
    ) { paddingValues ->
        if (state.isLoading) {
            Box(
                modifier = Modifier
                    .fillMaxSize()
                    .padding(paddingValues),
            ) {
                CircularProgressIndicator(modifier = Modifier.align(Alignment.Center))
            }
            return@Scaffold
        }

        CustomPullToRefresh(
            modifier = Modifier
                .fillMaxSize()
                .padding(paddingValues),
            isRefreshing = state.isRefreshing,
            onRefresh = {
                uiStateHolder.onLoadData(onStart = false)
            }
        ) {
            LazyColumn(
                modifier = Modifier.fillMaxSize(),
                contentPadding = PaddingValues(16.dp),
                verticalArrangement = Arrangement.spacedBy(8.dp)
            ) {

                items(
                    items = state.scripts,
                    key = { item -> item },
                ) {
                    ScriptItem(
                        modifier = Modifier.fillMaxWidth(),
                        name = it,
                        onPlayClick = uiStateHolder::onPlayScript,
                        onDeleteClick = uiStateHolder::onDeleteScript,
                    )
                }
            }
        }

        if (state.scriptNameToDelete.isNotEmpty()) {
            DeleteDialog(
                title = stringResource(R.string.script_delete_dialog_title),
                text = stringResource(
                    R.string.script_delete_dialog_text,
                    state.scriptNameToDelete,
                ),
                onDismiss = uiStateHolder::onDismissDeleteDialog,
                onConfirm = uiStateHolder::onConfirmDeleteScript,
            )
        }
    }
}

@Preview
@Composable
private fun MainUiLoadingScreenPreview() {
    ScriptsScreen(
        uiStateHolder = ScriptsScreenUiStateHolderPreview(scriptsUiStatePreview),
        snackbarHostState = remember { SnackbarHostState() }
    )
}