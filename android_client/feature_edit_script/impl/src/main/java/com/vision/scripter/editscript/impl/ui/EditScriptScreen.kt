package com.vision.scripter.editscript.impl.ui

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.imePadding
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.material3.AlertDialog
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Scaffold
import androidx.compose.material3.SnackbarHostState
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.runtime.Stable
import androidx.compose.runtime.remember
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.tooling.preview.Preview
import androidx.compose.ui.unit.dp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.vision.scripter.editscript.impl.R
import com.vision.scripter.ui.ProvideSnackbarHost
import com.vision.scripter.ui.R as CommonR

@Composable
internal fun EditScriptScreen(
    uiStateHolder: EditScriptUiStateHolder,
    snackbarHostState: SnackbarHostState,
) {
    val state = uiStateHolder.uiStateFlow.collectAsStateWithLifecycle(
        initialValue = EditScriptUiState(),
    ).value

    Scaffold(
        modifier = Modifier.fillMaxSize(),
        topBar = {
            EditScriptTopBar(
                name = state.name,
                onBackClick = uiStateHolder::onBackClicked,
                onSaveClick = uiStateHolder::onSaveClicked,
            )
        },
        snackbarHost = {
            ProvideSnackbarHost(snackbarHostState = snackbarHostState)
        },
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

        LazyColumn(
            modifier = Modifier
                .fillMaxSize()
                .padding(paddingValues)
                .imePadding(),
            contentPadding = PaddingValues(16.dp),
            verticalArrangement = Arrangement.spacedBy(8.dp),
        ) {
            item(key = "node_field") {
                ScriptEditField(
                    modifier = Modifier.fillMaxWidth(),
                    value = state.node,
                    label = stringResource(R.string.node_hint),
                    onValueChange = uiStateHolder::onNodeChanged,
                )
            }

            item(key = "next_node_field") {
                ScriptEditField(
                    modifier = Modifier.fillMaxWidth(),
                    value = state.nextNode,
                    label = stringResource(R.string.next_node_hint),
                    onValueChange = uiStateHolder::onNextNodeChanged,
                )
            }

            item(key = "timeout_field") {
                ScriptEditField(
                    modifier = Modifier.fillMaxWidth(),
                    value = state.timeout,
                    label = stringResource(R.string.timeout_hint),
                    keyboardType = KeyboardType.Number,
                    onValueChange = uiStateHolder::onTimeoutChanged,
                )
            }

            items(
                items = state.params,
                key = @Stable { param -> "param_${param.id}" },
            ) { param ->
                ParamCard(
                    modifier = Modifier.fillMaxWidth(),
                    param = param,
                    onDeleteClick = uiStateHolder::onDeleteParam,
                )
            }

            if (state.eventsCount > 0) {
                item(key = "events_card") {
                    EventsCard(
                        modifier = Modifier.fillMaxWidth(),
                        count = state.eventsCount,
                        onDeleteClick = uiStateHolder::onDeleteEvents,
                    )
                }
            }
        }
    }

    if (state.showSaveDialog) {
        SaveScriptDialog(
            name = state.name,
            onDismiss = uiStateHolder::onDismissDialog,
            onConfirm = uiStateHolder::onConfirmSave,
        )
    }
}

@Composable
private fun ScriptEditField(
    modifier: Modifier = Modifier,
    value: String,
    label: String,
    keyboardType: KeyboardType = KeyboardType.Text,
    onValueChange: (String) -> Unit,
) {
    OutlinedTextField(
        modifier = modifier,
        value = value,
        onValueChange = onValueChange,
        keyboardOptions = remember(keyboardType) { KeyboardOptions(keyboardType = keyboardType) },
        label = {
            Text(text = label)
        },
        singleLine = true,
    )
}

@Composable
private fun SaveScriptDialog(
    name: String,
    onDismiss: () -> Unit,
    onConfirm: () -> Unit,
) {
    AlertDialog(
        title = {
            Text(text = stringResource(R.string.save_dialog_title))
        },
        text = {
            Text(text = stringResource(R.string.save_dialog_text, name))
        },
        onDismissRequest = onDismiss,
        confirmButton = {
            TextButton(onClick = onConfirm) {
                Text(text = stringResource(CommonR.string.ok))
            }
        },
        dismissButton = {
            TextButton(onClick = onDismiss) {
                Text(text = stringResource(CommonR.string.cancel))
            }
        },
    )
}

@Preview
@Composable
private fun EditScriptScreenPreview() {
    EditScriptScreen(
        uiStateHolder = EditScriptUiStateHolderPreview(editScriptUiStatePreview),
        snackbarHostState = SnackbarHostState(),
    )
}

@Preview
@Composable
private fun EditScriptScreenLoadingPreview() {
    EditScriptScreen(
        uiStateHolder = EditScriptUiStateHolderPreview(EditScriptUiState()),
        snackbarHostState = SnackbarHostState(),
    )
}
