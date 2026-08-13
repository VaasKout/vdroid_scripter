package com.vision.scripter.library.ui

import com.vision.scripter.library.state.LibraryType
import com.vision.scripter.library.state.LibraryUiCommand
import com.vision.scripter.library.state.LibraryUiState
import com.vision.scripter.library.state.LibraryUiStateHolder
import com.vision.scripter.ui.CommandFlow
import kotlinx.collections.immutable.persistentListOf
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharedFlow

internal val libraryUiStatePreview = LibraryUiState(
    isLoading = false,
    images = persistentListOf(
        "x5_catalog_cart_icon",
        "tg_chat_send_button",
    ),
    actions = persistentListOf(
        "swipe_x5_catalog_1",
        "swipe_x5_catalog_2",
    ),
)

internal class LibraryUiStateHolderPreview(state: LibraryUiState) : LibraryUiStateHolder {
    override val uiStateFlow: SharedFlow<LibraryUiState> = MutableStateFlow(state)
    override val uiCommandsFlow: CommandFlow<LibraryUiCommand>
        get() = throw UnsupportedOperationException()

    override fun onLoadData(onStart: Boolean) {}
    override fun onDeleteItem(kind: LibraryType, name: String) {}
    override fun onDismiss() {}
    override fun onConfirmDelete() {}
}
