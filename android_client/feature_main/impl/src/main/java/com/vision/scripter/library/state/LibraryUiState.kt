package com.vision.scripter.library.state

import androidx.compose.runtime.Immutable
import com.vision.scripter.library.state.LibraryState.ItemToDelete
import kotlinx.collections.immutable.ImmutableList
import kotlinx.collections.immutable.persistentListOf

@Immutable
data class LibraryUiState(
    val isLoading: Boolean = true,
    val isRefreshing: Boolean = false,
    val images: ImmutableList<String> = persistentListOf(),
    val actions: ImmutableList<String> = persistentListOf(),
    val itemToDelete: ItemToDelete? = null,
)
