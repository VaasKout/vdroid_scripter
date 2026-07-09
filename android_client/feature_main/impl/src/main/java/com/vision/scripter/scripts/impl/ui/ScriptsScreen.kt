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
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.tooling.preview.Preview
import androidx.compose.ui.unit.dp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.vision.scripter.main.impl.R
import com.vision.scripter.scripts.impl.state.ScriptsUiState
import com.vision.scripter.scripts.impl.state.ScriptsUiStateHolder
import com.vision.scripter.ui.CustomPullToRefresh
import com.vision.scripter.ui.DeleteDialog

@Composable
internal fun ScriptsScreen(
    uiStateHolder: ScriptsUiStateHolder,
    paddingValues: PaddingValues,
) {
    val state = uiStateHolder.uiStateFlow.collectAsStateWithLifecycle(
        initialValue = ScriptsUiState(),
    ).value

    LaunchedEffect(Unit) {
        uiStateHolder.onLoadData(onStart = true)
    }

    if (state.isLoading) {
        Box(
            modifier = Modifier
                .fillMaxSize(),
        ) {
            CircularProgressIndicator(modifier = Modifier.align(Alignment.Center))
        }
        return
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

@Preview
@Composable
private fun MainUiLoadingScreenPreview() {
        ScriptsScreen(
            uiStateHolder = ScriptsScreenUiStateHolderPreview(scriptsUiStatePreview),
            paddingValues = PaddingValues(0.dp)
        )
}