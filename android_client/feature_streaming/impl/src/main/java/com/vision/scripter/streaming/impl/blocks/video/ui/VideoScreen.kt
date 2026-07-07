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
import com.vision.scripter.streaming.impl.screen.main.ui.StreamingPreviewScreenUiStateHolder
import com.vision.scripter.streaming.impl.screen.main.ui.StreamingUiState
import com.vision.scripter.streaming.impl.screen.main.ui.StreamingUiStateHolder
import com.vision.scripter.streaming.impl.screen.main.ui.streamingPreviewState

@Composable
fun VideoScreen(
    modifier: Modifier = Modifier,
    state: StreamingUiState,
    uiStateHolder: StreamingUiStateHolder,
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
                    onScriptModeClick = uiStateHolder::onScriptModeClicked,
                    onKeyboardClick = uiStateHolder::onKeyboardClicked,
                    onCvModeClick = uiStateHolder::onCvModeClicked,
                    onTextModeClick = uiStateHolder::onTextModeClicked,
                    onExpandClick = uiStateHolder::onExpandClicked,
                    onExitClick = uiStateHolder::exit,
                )
            }

            is MenuState.SelectingCV -> {
                SelectingTemplateMenu(
                    modifier = Modifier
                        .align(Alignment.TopEnd)
                        .padding(top = 128.dp),
                    flags = state.menuState.flags,
                    onCvModeClick = uiStateHolder::onCvModeClicked,
                    onSaveClick = uiStateHolder::onSaveClicked,
                    onBackClick = uiStateHolder::onCancelClicked,
                )
            }

            is MenuState.SelectingText -> {
                SelectingTextMenu(
                    modifier = Modifier
                        .align(Alignment.TopEnd)
                        .padding(top = 128.dp),
                    flags = state.menuState.flags,
                    onTextModeClick = uiStateHolder::onTextModeClicked,
                    onSaveClick = uiStateHolder::onSaveClicked,
                    onBackClick = uiStateHolder::onCancelClicked,
                )
            }

            is MenuState.Recording -> {
                ScriptMenu(
                    modifier = Modifier
                        .align(Alignment.TopEnd)
                        .padding(top = 128.dp),
                    menuState = state.menuState,
                    onRecordingClick = uiStateHolder::onRecordingClicked,
                    onCvModeClick = uiStateHolder::onCvModeClicked,
                    onTextModeClick = uiStateHolder::onTextModeClicked,
                    onKeyboardClick = uiStateHolder::onKeyboardClicked,
                    onTimeoutClick = uiStateHolder::onTimeoutClicked,
                    onSaveClick = uiStateHolder::onSaveClicked,
                    onCancelClick = uiStateHolder::onCancelClicked,
                )
            }

            is MenuState.Keyboard -> {
                KeyboardMenu(
                    modifier = Modifier
                        .align(Alignment.TopEnd)
                        .padding(top = 128.dp),
                    menuState = state.menuState,
                    onKeyboardInitClick = uiStateHolder::onKeyboardInitClicked,
                    onKeyboardRecordingClick = uiStateHolder::onRecordingClicked,
                    onKeyboardEdit = uiStateHolder::onKeyboardEdited,
                    onSaveClick = uiStateHolder::onSaveClicked,
                    onCancelClick = uiStateHolder::onCancelClicked,
                )
            }
        }

        when (state.dialogState) {
            DialogState.NONE -> {}

            DialogState.RECORD -> {
                RecordDialog(
                    onSaveRecordName = uiStateHolder::onSavedRecordName,
                    onDismiss = uiStateHolder::onDialogDismissed,
                )
            }

            DialogState.TEXT -> {
                TextToFindDialog(
                    onTryToFindText = uiStateHolder::onTryToFindText,
                    onDismiss = uiStateHolder::onDialogDismissed,
                )
            }

            DialogState.KEYBOARD -> {
                KeyboardDialog(
                    onSaveLocale = uiStateHolder::onSaveLocale,
                    onDismiss = uiStateHolder::onDialogDismissed,
                )
            }

            DialogState.EDIT_KEYBOARD -> {
                EditKeyboardDialog(
                    oldKey = (state.menuState as? MenuState.Keyboard)?.oldKey.orEmpty(),
                    onSave = uiStateHolder::onEditKeyboardButtonSaved,
                    onDismiss = uiStateHolder::onDialogDismissed,
                )
            }

            DialogState.TIMEOUT -> {
                TimeoutDialog(
                    initialTimeout = state.recordTimeout,
                    onSave = uiStateHolder::onTimeoutSaved,
                    onDismiss = uiStateHolder::onDialogDismissed,
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
    )
}