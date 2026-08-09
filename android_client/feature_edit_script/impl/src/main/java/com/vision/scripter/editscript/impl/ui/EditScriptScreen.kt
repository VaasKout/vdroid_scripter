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
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Scaffold
import androidx.compose.material3.SnackbarHostState
import androidx.compose.material3.Text
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
import com.vision.scripter.ui.CommonDialog
import com.vision.scripter.ui.DeleteDialog
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
                deleteMode = state.deleteMode,
                onBackClick = uiStateHolder::onBackClicked,
                onActionClick = uiStateHolder::onTopbarActionClicked,
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
            item(key = "location_field") {
                ScriptEditField(
                    modifier = Modifier.fillMaxWidth(),
                    value = state.location,
                    label = stringResource(R.string.location_hint),
                    onValueChange = uiStateHolder::onLocationChanged,
                )
            }

            item(key = "next_location_field") {
                ScriptEditField(
                    modifier = Modifier.fillMaxWidth(),
                    value = state.nextLocation,
                    label = stringResource(R.string.next_location_hint),
                    onValueChange = uiStateHolder::onNextLocationChanged,
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

            if (state.params.isNotEmpty()) {
                item(key = "params_title") {
                    SectionTitle(
                        modifier = Modifier.fillMaxWidth(),
                        text = stringResource(R.string.params_title),
                    )
                }
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

            item(key = "events_title") {
                SectionTitle(
                    modifier = Modifier.fillMaxWidth(),
                    text = stringResource(R.string.events_title, state.eventsCount),
                )
            }

            item(key = "delete_events_button") {
                DeleteEventsButton(
                    modifier = Modifier.fillMaxWidth(),
                    enabled = state.eventsCount > 0,
                    onClick = uiStateHolder::onDeleteEvents,
                )
            }
        }
    }

    if (state.showDialog) {
        Dialog(
            name = state.name,
            deleteMode = state.deleteMode,
            onDismiss = uiStateHolder::onDismissDialog,
            onConfirm = uiStateHolder::onConfirmDialog,
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
private fun Dialog(
    name: String,
    deleteMode: Boolean,
    onDismiss: () -> Unit,
    onConfirm: () -> Unit,
) {
    if (deleteMode) {
        DeleteDialog(
            title = stringResource(R.string.save_dialog_title),
            text = stringResource(R.string.delete_dialog_text, name),
            onDismiss = onDismiss,
            onConfirm = onConfirm,
        )
        return
    }

    CommonDialog(
        title = stringResource(R.string.save_dialog_title),
        text = stringResource(R.string.save_dialog_text, name),
        confirmButtonText = stringResource(CommonR.string.ok),
        dismissButtonText = stringResource(CommonR.string.cancel),
        onDismiss = onDismiss,
        onConfirm = onConfirm,
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
