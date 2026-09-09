package com.vision.scripter.library.ui

import com.vision.scripter.library.state.LibraryType
import com.vision.scripter.library.state.LibraryUiCommand
import com.vision.scripter.library.state.LibraryUiState
import com.vision.scripter.library.state.LibraryUiStateHolder
import com.vision.scripter.ui.CommandFlow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharedFlow

internal val libraryUiStatePreview = LibraryUiState(
    isLoading = false,
    imagesCount = 12,
    actionsCount = 4,
    routesCount = 2,
)

internal class LibraryUiStateHolderPreview(state: LibraryUiState) : LibraryUiStateHolder {
    override val uiStateFlow: SharedFlow<LibraryUiState> = MutableStateFlow(state)
    override val uiCommandsFlow: CommandFlow<LibraryUiCommand>
        get() = throw UnsupportedOperationException()

    override fun onLoadData(onStart: Boolean) {}
    override fun onCardClicked(type: LibraryType) {}
}
