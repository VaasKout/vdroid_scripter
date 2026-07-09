package com.vision.scripter.scripts.impl.state

import androidx.compose.runtime.Immutable
import kotlinx.collections.immutable.ImmutableList
import kotlinx.collections.immutable.persistentListOf

@Immutable
data class ScriptsUiState(
    val isLoading: Boolean = true,
    val isRefreshing: Boolean = false,
    val scripts: ImmutableList<String> = persistentListOf(),
    val scriptNameToDelete: String = "",
)