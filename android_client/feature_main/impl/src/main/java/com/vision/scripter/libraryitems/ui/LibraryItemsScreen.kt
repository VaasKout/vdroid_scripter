package com.vision.scripter.libraryitems.ui

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Scaffold
import androidx.compose.material3.SnackbarHostState
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.tooling.preview.Preview
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.vision.scripter.library.state.LibraryType
import com.vision.scripter.libraryitems.ui.items.DevicePickerSheet
import com.vision.scripter.libraryitems.ui.items.LibraryItem
import com.vision.scripter.main.impl.R
import com.vision.scripter.ui.CustomPullToRefresh
import com.vision.scripter.ui.DeleteDialog
import com.vision.scripter.ui.ProvideSnackbarHost
import com.vision.scripter.ui.TopBar
import com.vision.scripter.ui.customClickable

@Composable
internal fun LibraryItemsScreen(
    uiStateHolder: LibraryItemsUiStateHolder,
    type: LibraryType,
    snackbarHostState: SnackbarHostState,
    onBack: () -> Unit,
) {
    val state = uiStateHolder.uiStateFlow.collectAsStateWithLifecycle().value

    LaunchedEffect(type) {
        uiStateHolder.init(type)
    }

    Scaffold(
        modifier = Modifier.fillMaxSize(),
        topBar = {
            LibraryItemsTopBar(
                title = stringResource(type.titleRes()),
                onBack = onBack,
            )
        },
        snackbarHost = { ProvideSnackbarHost(snackbarHostState) },
    ) { paddingValues ->
        LibraryItemsContent(
            state = state,
            uiStateHolder = uiStateHolder,
            paddingValues = paddingValues,
        )
    }

    val itemToDelete = state.itemToDelete
    if (itemToDelete != null) {
        DeleteDialog(
            title = stringResource(R.string.delete_dialog_title),
            text = stringResource(type.deleteTextRes(), itemToDelete),
            onDismiss = uiStateHolder::onDismissDelete,
            onConfirm = uiStateHolder::onConfirmDelete,
        )
    }

    val picker = state.picker
    if (picker != null) {
        DevicePickerSheet(
            picker = picker,
            onDeviceSelected = uiStateHolder::onDeviceSelected,
            onPlayClick = uiStateHolder::onPickerPlayClicked,
            onDismiss = uiStateHolder::onPickerDismissed,
        )
    }
}

@Composable
private fun LibraryItemsContent(
    state: LibraryItemsUiState,
    uiStateHolder: LibraryItemsUiStateHolder,
    paddingValues: PaddingValues,
) {
    if (state.isLoading) {
        Box(
            modifier = Modifier
                .fillMaxSize()
                .padding(paddingValues),
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
        onRefresh = { uiStateHolder.onLoadData(onStart = false) },
    ) {
        LazyColumn(
            modifier = Modifier.fillMaxSize(),
            contentPadding = PaddingValues(16.dp),
            verticalArrangement = Arrangement.spacedBy(8.dp),
        ) {
            items(
                items = state.items,
                key = { item -> item.name },
            ) {
                LibraryItem(
                    modifier = Modifier.fillMaxWidth(),
                    item = it,
                    onPlayClick = uiStateHolder::onPlayClicked,
                    onDeleteClick = uiStateHolder::onDeleteItem,
                    onRunDismiss = uiStateHolder::onRunDismissed,
                )
            }
        }
    }
}

@Composable
private fun LibraryItemsTopBar(
    title: String,
    onBack: () -> Unit,
) {
    TopBar(
        startContent = {
            Icon(
                modifier = Modifier.customClickable(onClick = onBack),
                imageVector = Icons.AutoMirrored.Filled.ArrowBack,
                tint = MaterialTheme.colorScheme.onPrimary,
                contentDescription = null,
            )
            Text(
                modifier = Modifier.padding(start = 16.dp),
                text = title,
                style = TextStyle(
                    color = MaterialTheme.colorScheme.onPrimary,
                    fontSize = 20.sp,
                    fontWeight = FontWeight.Bold,
                )
            )
        },
    )
}

private fun LibraryType.titleRes(): Int = when (this) {
    LibraryType.IMAGES -> R.string.images
    LibraryType.ACTIONS -> R.string.actions
    LibraryType.ROUTES -> R.string.routes
}

private fun LibraryType.deleteTextRes(): Int = when (this) {
    LibraryType.IMAGES -> R.string.image_delete_dialog_text
    LibraryType.ACTIONS -> R.string.action_delete_dialog_text
    LibraryType.ROUTES -> R.string.route_delete_dialog_text
}

@Preview
@Composable
private fun LibraryItemsScreenPreview() {
    LibraryItemsScreen(
        uiStateHolder = LibraryItemsUiStateHolderPreview(libraryItemsUiStatePreview),
        type = LibraryType.ACTIONS,
        snackbarHostState = SnackbarHostState(),
        onBack = {},
    )
}
