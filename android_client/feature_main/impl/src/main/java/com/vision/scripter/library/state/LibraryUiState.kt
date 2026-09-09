package com.vision.scripter.library.state

import androidx.compose.runtime.Immutable

@Immutable
data class LibraryUiState(
    val isLoading: Boolean = true,
    val isRefreshing: Boolean = false,
    val imagesCount: Int = 0,
    val actionsCount: Int = 0,
    val routesCount: Int = 0,
)
