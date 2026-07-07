package com.vision.scripter.streaming.impl.blocks.video.ui

import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.tooling.preview.Preview
import androidx.compose.ui.unit.dp
import com.vision.scripter.streaming.impl.blocks.menu.state.DialogState
import com.vision.scripter.streaming.impl.blocks.menu.state.MenuUiStateHolder
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
import com.vision.scripter.streaming.impl.screen.main.state.MenuState
import com.vision.scripter.streaming.impl.screen.main.ui.StreamingPreviewMenuUiStateHolder
import com.vision.scripter.streaming.impl.screen.main.ui.StreamingPreviewScreenUiStateHolder
import com.vision.scripter.streaming.impl.screen.main.ui.StreamingUiState
import com.vision.scripter.streaming.impl.screen.main.ui.StreamingUiStateHolder
import com.vision.scripter.streaming.impl.screen.main.ui.streamingPreviewState

@Composable
fun VideoScreen(
    modifier: Modifier = Modifier,
    state: StreamingUiState,
    uiStateHolder: StreamingUiStateHolder,
    menuUiStateHolder: MenuUiStateHolder,
) {
    Box(modifier = modifier) {
        VideoSurface(
            modifier = Modifier.fillMaxSize(),
            onSurfaceCreated = uiStateHolder::onVideoSurfaceCreated,
            onSurfaceDestroyed = uiStateHolder::onVideoSurfaceDestroyed,
            onTouch = uiStateHolder::onTouchEvent,
        )

        RectanglesCanvas(
            modifier = Modifier.fillMaxSize(),
            cvRectangles = state.rectangles,
            selectedRectangles = state.selectedRectangles,
            keyboardButtons = state.keyboardButtons,
        )

        when (state.menuState) {
            is MenuState.Usual -> {
                UsualMenu(
                    modifier = Modifier
                        .align(Alignment.TopEnd)
                        .padding(top = 128.dp),
                    menuState = state.menuState,
                    onScriptModeClick = menuUiStateHolder::onScriptModeClicked,
                    onKeyboardClick = menuUiStateHolder::onKeyboardClicked,
                    onCvModeClick = menuUiStateHolder::onCvModeClicked,
                    onTextModeClick = menuUiStateHolder::onTextModeClicked,
                    onExpandClick = menuUiStateHolder::onExpandClicked,
                    onExitClick = menuUiStateHolder::onExitClicked,
                )
            }

            is MenuState.SelectingCV -> {
                SelectingTemplateMenu(
                    modifier = Modifier
                        .align(Alignment.TopEnd)
                        .padding(top = 128.dp),
                    flags = state.menuState.flags,
                    onCvModeClick = menuUiStateHolder::onCvModeClicked,
                    onSaveClick = menuUiStateHolder::onSaveClicked,
                    onBackClick = menuUiStateHolder::onCancelClicked,
                )
            }

            is MenuState.SelectingText -> {
                SelectingTextMenu(
                    modifier = Modifier
                        .align(Alignment.TopEnd)
                        .padding(top = 128.dp),
                    flags = state.menuState.flags,
                    onTextModeClick = menuUiStateHolder::onTextModeClicked,
                    onSaveClick = menuUiStateHolder::onSaveClicked,
                    onBackClick = menuUiStateHolder::onCancelClicked,
                )
            }

            is MenuState.Recording -> {
                ScriptMenu(
                    modifier = Modifier
                        .align(Alignment.TopEnd)
                        .padding(top = 128.dp),
                    menuState = state.menuState,
                    onRecordingClick = menuUiStateHolder::onRecordingClicked,
                    onCvModeClick = menuUiStateHolder::onCvModeClicked,
                    onTextModeClick = menuUiStateHolder::onTextModeClicked,
                    onKeyboardClick = menuUiStateHolder::onKeyboardClicked,
                    onTimeoutClick = menuUiStateHolder::onTimeoutClicked,
                    onSaveClick = menuUiStateHolder::onSaveClicked,
                    onCancelClick = menuUiStateHolder::onCancelClicked,
                )
            }

            is MenuState.Keyboard -> {
                KeyboardMenu(
                    modifier = Modifier
                        .align(Alignment.TopEnd)
                        .padding(top = 128.dp),
                    menuState = state.menuState,
                    onKeyboardInitClick = menuUiStateHolder::onKeyboardInitClicked,
                    onKeyboardRecordingClick = menuUiStateHolder::onRecordingClicked,
                    onKeyboardEdit = menuUiStateHolder::onKeyboardEdited,
                    onSaveClick = menuUiStateHolder::onSaveClicked,
                    onCancelClick = menuUiStateHolder::onCancelClicked,
                )
            }
        }

        when (state.dialogState) {
            DialogState.NONE -> {}

            DialogState.RECORD -> {
                RecordDialog(
                    onSaveRecordName = menuUiStateHolder::onSavedRecordName,
                    onDismiss = menuUiStateHolder::onDialogDismissed,
                )
            }

            DialogState.TEXT -> {
                TextToFindDialog(
                    onTryToFindText = menuUiStateHolder::onTryToFindText,
                    onDismiss = menuUiStateHolder::onDialogDismissed,
                )
            }

            DialogState.KEYBOARD -> {
                KeyboardDialog(
                    onSaveLocale = menuUiStateHolder::onSaveLocale,
                    onDismiss = menuUiStateHolder::onDialogDismissed,
                )
            }

            DialogState.EDIT_KEYBOARD -> {
                EditKeyboardDialog(
                    oldKey = (state.menuState as? MenuState.Keyboard)?.oldKey.orEmpty(),
                    onSave = menuUiStateHolder::onEditKeyboardButtonSaved,
                    onDismiss = menuUiStateHolder::onDialogDismissed,
                )
            }

            DialogState.TIMEOUT -> {
                TimeoutDialog(
                    initialTimeout = state.recordTimeout,
                    onSave = menuUiStateHolder::onTimeoutSaved,
                    onDismiss = menuUiStateHolder::onDialogDismissed,
                )
            }
        }
    }
}

@Preview
@Composable
private fun VideoScreenPreview() {
    VideoScreen(
        state = streamingPreviewState,
        uiStateHolder = StreamingPreviewScreenUiStateHolder(streamingPreviewState),
        menuUiStateHolder = StreamingPreviewMenuUiStateHolder(),
    )
}
