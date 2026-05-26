package com.vision.scripter.streaming.impl.ui

import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.tooling.preview.Preview
import androidx.compose.ui.unit.dp
import com.vision.scripter.streaming.impl.state.DialogState
import com.vision.scripter.streaming.impl.state.MenuState
import com.vision.scripter.streaming.impl.ui.items.KeyboardDialog
import com.vision.scripter.streaming.impl.ui.items.RecordDialog
import com.vision.scripter.streaming.impl.ui.items.RectanglesCanvas
import com.vision.scripter.streaming.impl.ui.items.TextToFindDialog
import com.vision.scripter.streaming.impl.ui.items.VideoSurface
import com.vision.scripter.streaming.impl.ui.items.menu.KeyboardMenu
import com.vision.scripter.streaming.impl.ui.items.menu.ScriptMenu
import com.vision.scripter.streaming.impl.ui.items.menu.SelectingTemplateMenu
import com.vision.scripter.streaming.impl.ui.items.menu.SelectingTextMenu
import com.vision.scripter.streaming.impl.ui.items.menu.UsualMenu

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
                )
            }

            is MenuState.SelectingCV -> {
                SelectingTemplateMenu(
                    modifier = Modifier
                        .align(Alignment.TopEnd)
                        .padding(top = 128.dp),
                    mode = state.menuState.selectMode,
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
                    mode = state.menuState.selectMode,
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