package com.vision.scripter.streaming.impl.blocks.menu.ui

import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.tooling.preview.Preview
import androidx.hilt.lifecycle.viewmodel.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.navigation.NavController
import com.vision.scripter.streaming.impl.blocks.menu.commandobservers.MenuCommandObserver
import com.vision.scripter.streaming.impl.blocks.menu.state.DialogState
import com.vision.scripter.streaming.impl.blocks.menu.state.MenuState
import com.vision.scripter.streaming.impl.blocks.menu.state.MenuViewModel
import com.vision.scripter.streaming.impl.blocks.menu.ui.dialogs.EditKeyboardDialog
import com.vision.scripter.streaming.impl.blocks.menu.ui.dialogs.KeyboardDialog
import com.vision.scripter.streaming.impl.blocks.menu.ui.dialogs.RecordDialog
import com.vision.scripter.streaming.impl.blocks.menu.ui.dialogs.TextToFindDialog
import com.vision.scripter.streaming.impl.blocks.menu.ui.dialogs.TimeoutDialog
import com.vision.scripter.streaming.impl.blocks.menu.ui.menu.KeyboardMenu
import com.vision.scripter.streaming.impl.blocks.menu.ui.menu.ScriptMenu
import com.vision.scripter.streaming.impl.blocks.menu.ui.menu.SelectingTemplateMenu
import com.vision.scripter.streaming.impl.blocks.menu.ui.menu.SelectingTextMenu
import com.vision.scripter.streaming.impl.blocks.menu.ui.menu.UsualMenu
import com.vision.scripter.streaming.impl.screen.main.state.DEFAULT_TIMEOUT

@Composable
fun MenuBlock(
    modifier: Modifier = Modifier,
    navController: NavController,
) {
    val menuUiStateHolder = hiltViewModel<MenuViewModel>()
    val state = menuUiStateHolder.uiStateFlow.collectAsStateWithLifecycle().value

    MenuCommandObserver(
        uiStateHolder = menuUiStateHolder,
        navController = navController,
    )

    MenuContent(
        modifier = modifier,
        state = state,
        uiStateHolder = menuUiStateHolder,
    )
}

@Composable
private fun MenuContent(
    modifier: Modifier = Modifier,
    state: MenuUiState,
    uiStateHolder: MenuUiStateHolder,
) {
    when (state.menuState) {
        is MenuState.Usual -> {
            UsualMenu(
                modifier = modifier,
                menuState = state.menuState,
                uiStateHolder = uiStateHolder,
            )
        }

        is MenuState.SelectingCV -> {
            SelectingTemplateMenu(
                modifier = modifier,
                cvMode = state.menuState.cvMode,
                uiStateHolder = uiStateHolder,
            )
        }

        is MenuState.SelectingText -> {
            SelectingTextMenu(
                modifier = modifier,
                uiStateHolder = uiStateHolder,
            )
        }

        is MenuState.Recording -> {
            ScriptMenu(
                modifier = modifier,
                menuState = state.menuState,
                uiStateHolder = uiStateHolder,
            )
        }

        is MenuState.Keyboard -> {
            KeyboardMenu(
                modifier = modifier,
                menuState = state.menuState,
                uiStateHolder = uiStateHolder,
            )
        }
    }

    when (state.dialogState) {
        is DialogState.None -> {}

        is DialogState.Record -> {
            RecordDialog(
                onSaveRecordName = uiStateHolder::onSavedRecordName,
                onDismiss = uiStateHolder::onDialogDismissed,
            )
        }

        is DialogState.Text -> {
            TextToFindDialog(
                onTryToFindText = uiStateHolder::onTryToFindText,
                onDismiss = uiStateHolder::onDialogDismissed,
            )
        }

        is DialogState.Keyboard -> {
            KeyboardDialog(
                onSaveLocale = uiStateHolder::onSaveLocale,
                onDismiss = uiStateHolder::onDialogDismissed,
            )
        }

        is DialogState.EditKeyboard -> {
            EditKeyboardDialog(
                oldKey = state.dialogState.oldKey,
                onSave = uiStateHolder::onEditKeyboardButtonSaved,
                onDismiss = uiStateHolder::onDialogDismissed,
            )
        }

        is DialogState.Timeout -> {
            TimeoutDialog(
                initialTimeout = (state.menuState as? MenuState.Recording)?.recordTimeout
                    ?: DEFAULT_TIMEOUT,
                onSave = uiStateHolder::onTimeoutSaved,
                onDismiss = uiStateHolder::onDialogDismissed,
            )
        }
    }
}

@Preview
@Composable
private fun MenuUiPreview() {
    MenuContent(
        state = MenuUiState(),
        uiStateHolder = MenuPreviewUiStateHolder(usualMenuPreviewUiState),
    )
}
