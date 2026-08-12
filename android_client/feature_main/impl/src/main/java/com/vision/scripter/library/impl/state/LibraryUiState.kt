package com.vision.scripter.library.impl.state

import androidx.compose.runtime.Immutable
import kotlinx.collections.immutable.ImmutableList
import kotlinx.collections.immutable.persistentListOf

@Immutable
data class LibraryUiState(
    val isLoading: Boolean = true,
    val isRefreshing: Boolean = false,
    val images: ImmutableList<String> = persistentListOf(),
    val actions: ImmutableList<String> = persistentListOf(),
    val itemToDelete: LibraryItemToDelete? = null,
)

@Immutable
data class LibraryItemToDelete(
    val kind: LibraryKind,
    val name: String,
)
