package com.vision.scripter.streaming.impl.ui.items.menu

import androidx.compose.foundation.layout.ColumnScope
import androidx.compose.foundation.layout.size
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Keyboard
import androidx.compose.material.icons.filled.TextFields
import androidx.compose.material.icons.filled.Visibility
import androidx.compose.material.icons.filled.VisibilityOff
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.unit.dp
import com.vision.scripter.streaming.impl.state.CvSelectMode
import com.vision.scripter.ui.customClickable

@Composable
fun ColumnScope.BaseMenuIcons(
    visibilitySelectMode: CvSelectMode,
    textSelectMode: CvSelectMode,
    keyboardHighlighted: Boolean,
    onCvModeClick: () -> Unit,
    onTextClick: () -> Unit,
    onKeyboardClick: () -> Unit,
) {
    Icon(
        modifier = Modifier
            .size(32.dp)
            .customClickable(onClick = onCvModeClick),
        imageVector = when (visibilitySelectMode) {
            CvSelectMode.VISIBLE,
            CvSelectMode.APPLY_EVENT -> Icons.Filled.Visibility

            CvSelectMode.NONE -> Icons.Filled.VisibilityOff
        },
        tint = visibilitySelectMode.tint(),
        contentDescription = "",
    )

    Icon(
        modifier = Modifier
            .size(32.dp)
            .customClickable(onClick = onTextClick),
        imageVector = Icons.Filled.TextFields,
        tint = textSelectMode.tint(),
        contentDescription = "",
    )

    Icon(
        modifier = Modifier
            .size(32.dp)
            .customClickable(onClick = onKeyboardClick),
        imageVector = Icons.Filled.Keyboard,
        tint = if (keyboardHighlighted) Color.Red else MaterialTheme.colorScheme.onSurface,
        contentDescription = "",
    )
}

@Composable
private fun CvSelectMode.tint(): Color = when (this) {
    CvSelectMode.VISIBLE -> Color.Blue
    CvSelectMode.APPLY_EVENT -> Color.Red
    CvSelectMode.NONE -> MaterialTheme.colorScheme.onSurface
}
