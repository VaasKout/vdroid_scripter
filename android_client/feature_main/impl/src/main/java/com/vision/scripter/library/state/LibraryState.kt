package com.vision.scripter.library.state

import com.vision.scripter.ui.states.LoadingState

enum class LibraryType {
    IMAGES,
    ACTIONS,
    ROUTES,
}

data class LibraryState(
    val loadingState: LoadingState = LoadingState.LoadingOnStart,
    val images: List<String> = listOf(),
    val actions: List<String> = listOf(),
    val routes: List<String> = listOf(),
)
