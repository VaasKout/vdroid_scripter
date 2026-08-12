package com.vision.scripter.streaming.impl.blocks.menu.ui

import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.tooling.preview.Preview
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.navigation.NavController
import com.vision.scripter.streaming.impl.blocks.menu.commandobservers.MenuCommandObserver
import com.vision.scripter.streaming.impl.blocks.menu.state.DialogState
import com.vision.scripter.streaming.impl.blocks.menu.state.MenuType
import com.vision.scripter.streaming.impl.blocks.menu.ui.dialogs.AddItemDialog
import com.vision.scripter.streaming.impl.blocks.menu.ui.dialogs.EditKeyboardDialog
import com.vision.scripter.streaming.impl.blocks.menu.ui.dialogs.KeyboardDialog
import com.vision.scripter.streaming.impl.blocks.menu.ui.dialogs.TextToFindDialog
import com.vision.scripter.streaming.impl.blocks.menu.ui.menu.CustomActionMenu
import com.vision.scripter.streaming.impl.blocks.menu.ui.menu.KeyboardMenu
import com.vision.scripter.streaming.impl.blocks.menu.ui.menu.SelectingTemplateMenu
import com.vision.scripter.streaming.impl.blocks.menu.ui.menu.UsualMenu

@Composable
fun MenuBlock(
    modifier: Modifier = Modifier,
    navController: NavController,
    uiStateHolder: MenuUiStateHolder,
) {
    val state = uiStateHolder.uiStateFlow.collectAsStateWithLifecycle().value
    MenuCommandObserver(
        uiStateHolder = uiStateHolder,
        navController = navController,
    )

    MenuContent(
        modifier = modifier,
        state = state,
        uiStateHolder = uiStateHolder,
    )
}

@Composable
private fun MenuContent(
    modifier: Modifier = Modifier,
    state: MenuUiState,
    uiStateHolder: MenuUiStateHolder,
) {
    when (state.menuType) {
        is MenuType.Usual -> {
            UsualMenu(
                modifier = modifier,
                menuType = state.menuType,
                uiStateHolder = uiStateHolder,
            )
        }

        is MenuType.SelectingCV -> {
            SelectingTemplateMenu(
                modifier = modifier,
                cvMode = state.menuType.localCvMode,
                uiStateHolder = uiStateHolder,
            )
        }

        is MenuType.CustomAction -> {
            CustomActionMenu(
                modifier = modifier,
                menuType = state.menuType,
                uiStateHolder = uiStateHolder,
            )
        }

        is MenuType.Keyboard -> {
            KeyboardMenu(
                modifier = modifier,
                menuType = state.menuType,
                uiStateHolder = uiStateHolder,
            )
        }
    }

    when (state.dialogState) {
        is DialogState.None -> {}

        is DialogState.AddItem -> {
            AddItemDialog(
                onConfirm = uiStateHolder::onAddItemConfirmed,
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
