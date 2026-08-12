package com.vision.scripter.library.impl.state

import com.vision.scripter.ui.states.LoadingState

enum class LibraryKind {
    IMAGES,
    ACTIONS,
}

data class LibraryState(
    val loadingState: LoadingState = LoadingState.LoadingOnStart,
    val images: List<String> = listOf(),
    val actions: List<String> = listOf(),
    val itemToDelete: ItemToDelete? = null,
) {
    data class ItemToDelete(
        val kind: LibraryKind,
        val name: String,
    )
}
