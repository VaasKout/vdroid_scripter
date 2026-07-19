package com.vision.scripter.streaming.impl.blocks.menu.ui.menu

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Check
import androidx.compose.material.icons.filled.Close
import androidx.compose.material.icons.filled.RadioButtonChecked
import androidx.compose.material.icons.filled.Timer
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.tooling.preview.Preview
import androidx.compose.ui.unit.dp
import com.vision.scripter.streaming.impl.screen.main.state.MenuState
import com.vision.scripter.streaming.impl.screen.main.state.TEMPLATE
import com.vision.scripter.streaming.impl.screen.main.state.TEXT
import com.vision.scripter.streaming.impl.screen.main.state.TYPE_TEXT
import com.vision.scripter.ui.customClickable
import kotlinx.collections.immutable.ImmutableSet
import kotlinx.collections.immutable.persistentSetOf

@Composable
fun ScriptMenu(
    modifier: Modifier = Modifier,
    menuState: MenuState.Recording,
    paramKeys: ImmutableSet<String>,
    onRecordingClick: () -> Unit,
    onCvModeClick: () -> Unit,
    onTextModeClick: () -> Unit,
    onKeyboardClick: () -> Unit,
    onTimeoutClick: () -> Unit,
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

        BaseMenuIcons(
            paramKeys = paramKeys,
            onCvModeClick = onCvModeClick,
            onTextClick = onTextModeClick,
            onKeyboardClick = onKeyboardClick,
        )

        Icon(
            modifier = Modifier
                .size(32.dp)
                .customClickable(onClick = onTimeoutClick),
            imageVector = Icons.Filled.Timer,
            tint =
                if (menuState.customTimeout) Color.Red
                else MaterialTheme.colorScheme.onSurface,
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
        menuState = MenuState.Recording(),
        paramKeys = persistentSetOf(),
        onRecordingClick = {},
        onCvModeClick = {},
        onTextModeClick = {},
        onKeyboardClick = {},
        onTimeoutClick = {},
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
            customTimeout = true,
        ),
        paramKeys = persistentSetOf(TEMPLATE, TEXT, TYPE_TEXT),
        onRecordingClick = {},
        onCvModeClick = {},
        onTextModeClick = {},
        onKeyboardClick = {},
        onTimeoutClick = {},
        onSaveClick = {},
        onCancelClick = {},
    )
}
