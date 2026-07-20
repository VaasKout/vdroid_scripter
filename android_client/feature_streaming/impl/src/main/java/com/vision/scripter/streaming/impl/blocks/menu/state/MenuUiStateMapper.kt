package com.vision.scripter.streaming.impl.blocks.menu.state

import com.vision.scripter.streaming.impl.blocks.menu.ui.MenuUiState
import dagger.hilt.android.scopes.ViewModelScoped
import javax.inject.Inject

@ViewModelScoped
class MenuUiStateMapper @Inject constructor() {
    fun map(
        state: MenuState,
        dialogState: DialogState,
    ): MenuUiState {
        return MenuUiState(
            menuState = state,
            dialogState = dialogState,
        )
    }
}