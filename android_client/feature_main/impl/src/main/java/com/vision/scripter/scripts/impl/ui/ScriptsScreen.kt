package com.vision.scripter.scripts.impl.ui

import androidx.activity.compose.BackHandler
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

    BackHandler(enabled = state.showScripts) {
        uiStateHolder.onBack()
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
            if (state.showScripts) {
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
                return@LazyColumn
            }

            items(
                items = state.nodes,
                key = { item -> item },
            ) {
                NodeItem(
                    modifier = Modifier.fillMaxWidth(),
                    name = it,
                    onClick = uiStateHolder::onNodeClick,
                    onDeleteClick = uiStateHolder::onDeleteNode,
                )
            }
        }
    }

    if (state.deleteDialog.show) {
        val text = if (state.deleteDialog.isNode) {
            stringResource(R.string.node_delete_dialog_text, state.deleteDialog.name)
        } else {
            stringResource(R.string.script_delete_dialog_text, state.deleteDialog.name)
        }
        DeleteDialog(
            title = stringResource(R.string.script_delete_dialog_title),
            text = text,
            onDismiss = uiStateHolder::onDismissDeleteDialog,
            onConfirm = uiStateHolder::onConfirmDelete,
        )
    }

    if (state.devicePickerData.showDevicePicker) {
        DevicePickerSheet(
            isLoading = state.devicePickerData.isDevicesLoading,
            devices = state.devicePickerData.devices,
            selectedSerial = state.devicePickerData.selectedSerial,
            onSelect = uiStateHolder::onSelectDevice,
            onConfirm = uiStateHolder::onConfirmRunScript,
            onDismiss = uiStateHolder::onDismissDevicePicker,
        )
    }
}

@Preview
@Composable
private fun ScriptsNodesScreenPreview() {
    ScriptsScreen(
        uiStateHolder = ScriptsScreenUiStateHolderPreview(scriptsUiStatePreview),
        paddingValues = PaddingValues(0.dp)
    )
}
