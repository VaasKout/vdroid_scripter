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
import androidx.compose.material.icons.filled.RadioButtonChecked
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.tooling.preview.Preview
import androidx.compose.ui.unit.dp
import com.vision.scripter.data.api.models.EVENT_ON_TEMPLATE
import com.vision.scripter.data.api.models.EVENT_ON_TEXT
import com.vision.scripter.data.api.models.TEMPLATE_IS_VISIBLE
import com.vision.scripter.data.api.models.TEXT_IS_VISIBLE
import com.vision.scripter.data.api.models.TYPE_TEXT
import com.vision.scripter.streaming.impl.state.MenuState
import com.vision.scripter.streaming.impl.state.hasFlag
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

        BaseMenuIcons(
            flags = menuState.flags,
            keyboardHighlighted = menuState.flags.hasFlag(TYPE_TEXT),
            onCvModeClick = onCvModeClick,
            onTextClick = onTextModeClick,
            onKeyboardClick = onKeyboardClick,
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
            flags = EVENT_ON_TEMPLATE or TEXT_IS_VISIBLE,
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
            flags = TEMPLATE_IS_VISIBLE or EVENT_ON_TEXT or TYPE_TEXT,
        ),
        onRecordingClick = {},
        onCvModeClick = {},
        onTextModeClick = {},
        onKeyboardClick = {},
        onSaveClick = {},
        onCancelClick = {},
    )
}
