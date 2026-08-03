package com.vision.scripter.editscript.impl.ui

import androidx.compose.runtime.Immutable
import kotlinx.collections.immutable.ImmutableList
import kotlinx.collections.immutable.persistentListOf

@Immutable
data class EditScriptUiState(
    val isLoading: Boolean = true,
    val name: String = "",
    val node: String = "",
    val nextNode: String = "",
    val timeout: String = "",
    val params: ImmutableList<ParamUiData> = persistentListOf(),
    val eventsCount: Int = 0,
    val showSaveDialog: Boolean = false,
)

@Immutable
data class ParamUiData(
    val id: Int = 0,
    val title: String = "",
    val locale: String = "",
)
