package com.vision.scripter.libraryitems.ui

import com.vision.scripter.library.state.LibraryType
import com.vision.scripter.ui.CommandFlow
import kotlinx.collections.immutable.persistentListOf
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow

internal val libraryItemsUiStatePreview = LibraryItemsUiState(
    type = LibraryType.ACTIONS,
    isLoading = false,
    items = persistentListOf(
        UiLibraryItem(name = "shop_catalog_swipe_1", canPlay = true),
        UiLibraryItem(
            name = "shop_checkout",
            canPlay = true,
            run = UiRun(
                statusText = "running step 2: tap on text Checkout",
                deviceLabel = "Pixel 6 (emulator-5554)",
                isError = false,
            ),
        ),
    ),
)

internal class LibraryItemsUiStateHolderPreview(
    state: LibraryItemsUiState,
) : LibraryItemsUiStateHolder {
    override val uiStateFlow: StateFlow<LibraryItemsUiState> = MutableStateFlow(state)
    override val uiCommandsFlow: CommandFlow<LibraryItemsUiCommand>
        get() = throw UnsupportedOperationException()

    override fun init(type: LibraryType) {}
    override fun onLoadData(onStart: Boolean) {}
    override fun onDeleteItem(name: String) {}
    override fun onDismissDelete() {}
    override fun onConfirmDelete() {}
    override fun onPlayClicked(name: String) {}
    override fun onDeviceChosen(serial: String) {}
    override fun onPickerDismissed() {}
    override fun onRunDismissed(name: String) {}
}
