package com.vision.scripter.streaming.impl.screen.main.ui

import androidx.compose.runtime.Immutable

@Immutable
data class StreamingUiState(
    val isLoading: Boolean = true,
    val isError: Boolean = false,
)
