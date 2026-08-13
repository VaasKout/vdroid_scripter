package com.vision.scripter.library.ui

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
import com.vision.scripter.library.state.LibraryType
import com.vision.scripter.library.state.LibraryUiState
import com.vision.scripter.library.state.LibraryUiStateHolder
import com.vision.scripter.main.impl.R
import com.vision.scripter.ui.CustomPullToRefresh
import com.vision.scripter.ui.DeleteDialog

@Composable
internal fun LibraryScreen(
    uiStateHolder: LibraryUiStateHolder,
    type: LibraryType,
    paddingValues: PaddingValues,
) {
    val state = uiStateHolder.uiStateFlow.collectAsStateWithLifecycle(
        initialValue = LibraryUiState(),
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
        val names = when (type) {
            LibraryType.IMAGES -> state.images
            LibraryType.ACTIONS -> state.actions
        }
        LazyColumn(
            modifier = Modifier.fillMaxSize(),
            contentPadding = PaddingValues(16.dp),
            verticalArrangement = Arrangement.spacedBy(8.dp)
        ) {
            items(
                items = names,
                key = { item -> item },
            ) {
                LibraryItem(
                    modifier = Modifier.fillMaxWidth(),
                    name = it,
                    onDeleteClick = { name ->
                        uiStateHolder.onDeleteItem(type, name)
                    },
                )
            }
        }
    }

    val itemToDelete = state.itemToDelete
    if (itemToDelete != null) {
        val text = when (itemToDelete.type) {
            LibraryType.IMAGES ->
                stringResource(R.string.image_delete_dialog_text, itemToDelete.name)

            LibraryType.ACTIONS ->
                stringResource(R.string.action_delete_dialog_text, itemToDelete.name)
        }
        DeleteDialog(
            title = stringResource(R.string.delete_dialog_title),
            text = text,
            onDismiss = uiStateHolder::onDismiss,
            onConfirm = uiStateHolder::onConfirmDelete,
        )
    }
}

@Preview
@Composable
private fun LibraryImagesScreenPreview() {
    LibraryScreen(
        uiStateHolder = LibraryUiStateHolderPreview(libraryUiStatePreview),
        type = LibraryType.IMAGES,
        paddingValues = PaddingValues(0.dp),
    )
}

@Preview
@Composable
private fun LibraryActionsScreenPreview() {
    LibraryScreen(
        uiStateHolder = LibraryUiStateHolderPreview(libraryUiStatePreview),
        type = LibraryType.ACTIONS,
        paddingValues = PaddingValues(0.dp),
    )
}
