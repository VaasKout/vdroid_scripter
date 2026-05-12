package com.vision.scripter.streaming.impl.ui.items.menu

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Check
import androidx.compose.material.icons.filled.Close
import androidx.compose.material.icons.filled.Keyboard
import androidx.compose.material.icons.filled.RadioButtonChecked
import androidx.compose.material.icons.filled.TextFields
import androidx.compose.material.icons.filled.Visibility
import androidx.compose.material.icons.filled.VisibilityOff
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.tooling.preview.Preview
import androidx.compose.ui.unit.dp
import com.vision.scripter.streaming.impl.state.CvSelectMode
import com.vision.scripter.streaming.impl.state.MenuState
import com.vision.scripter.ui.customClickable

@Composable
fun ScriptMenu(
    modifier: Modifier = Modifier,
    menuState: MenuState.Recording,
    onRecordingClick: () -> Unit,
    onCvModeClick: () -> Unit,
    onTextModeClick: () -> Unit,
    onKeyboardClick: () -> Unit,
    onSaveClick: () -> Unit,
    onCancelClick: () -> Unit,
) {
    Column(
        modifier = modifier
            .background(
                color = Color.White,
                shape = RoundedCornerShape(topStart = 8.dp, bottomStart = 8.dp),
            )
            .padding(horizontal = 4.dp),
        horizontalAlignment = Alignment.End,
        verticalArrangement = Arrangement.spacedBy(4.dp),
    ) {
        Icon(
            modifier = Modifier
                .size(32.dp)
                .customClickable(
                    onClick = onRecordingClick,
                ),
            imageVector = Icons.Filled.RadioButtonChecked,
            tint =
                if (menuState.controlRecording) Color.Red
                else MaterialTheme.colorScheme.onSurface,
            contentDescription = ""
        )

        Icon(
            modifier = Modifier
                .size(32.dp)
                .customClickable(onClick = onCvModeClick),
            imageVector = when (menuState.templateSelectMode) {
                CvSelectMode.VISIBLE,
                CvSelectMode.APPLY_EVENT -> Icons.Filled.Visibility

                else -> Icons.Filled.VisibilityOff
            },
            tint = when (menuState.templateSelectMode) {
                CvSelectMode.VISIBLE -> Color.Blue
                CvSelectMode.APPLY_EVENT -> Color.Red
                else -> MaterialTheme.colorScheme.onSurface
            },
            contentDescription = ""
        )

        Icon(
            modifier = Modifier
                .size(32.dp)
                .customClickable(onClick = onTextModeClick),
            imageVector = Icons.Filled.TextFields,
            tint = when (menuState.textSelectMode) {
                CvSelectMode.VISIBLE -> Color.Blue
                CvSelectMode.APPLY_EVENT -> Color.Red
                else -> MaterialTheme.colorScheme.onSurface
            },
            contentDescription = ""
        )

        Icon(
            modifier = Modifier
                .size(32.dp)
                .customClickable(onClick = onKeyboardClick),
            imageVector = Icons.Filled.Keyboard,
            tint = when {
                menuState.typeText -> Color.Red
                else -> MaterialTheme.colorScheme.onSurface
            },
            contentDescription = ""
        )

        Icon(
            modifier = Modifier
                .size(32.dp)
                .customClickable(onClick = onCancelClick),
            imageVector = Icons.Filled.Close,
            tint = MaterialTheme.colorScheme.onSurface,
            contentDescription = ""
        )

        Icon(
            modifier = Modifier
                .size(32.dp)
                .customClickable(onClick = onSaveClick),
            imageVector = Icons.Filled.Check,
            tint = Color.Green,
            contentDescription = ""
        )
    }
}

@Preview
@Composable
private fun ScriptMenuDefaultPreview() {
    ScriptMenu(
        menuState = MenuState.Recording(
            textSelectMode = CvSelectMode.VISIBLE,
            templateSelectMode = CvSelectMode.APPLY_EVENT,
        ),
        onRecordingClick = {},
        onCvModeClick = {},
        onTextModeClick = {},
        onKeyboardClick = {},
        onSaveClick = {},
        onCancelClick = {},
    )
}

@Preview
@Composable
private fun ScriptMenuRecordingPreview() {
    ScriptMenu(
        menuState = MenuState.Recording(
            controlRecording = true,
            textSelectMode = CvSelectMode.APPLY_EVENT,
            templateSelectMode = CvSelectMode.VISIBLE,
            typeText = true,
        ),
        onRecordingClick = {},
        onCvModeClick = {},
        onTextModeClick = {},
        onKeyboardClick = {},
        onSaveClick = {},
        onCancelClick = {},
    )
}
