package com.vision.scripter.streaming.impl.blocks.menu.ui.menu

import androidx.compose.foundation.layout.size
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Camera
import androidx.compose.material.icons.filled.Keyboard
import androidx.compose.material.icons.filled.TextFields
import androidx.compose.material.icons.filled.Visibility
import androidx.compose.material.icons.filled.VisibilityOff
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.unit.dp
import com.vision.scripter.streaming.impl.screen.main.state.TEMPLATE
import com.vision.scripter.streaming.impl.screen.main.state.TEXT
import com.vision.scripter.streaming.impl.screen.main.state.TYPE_TEXT
import com.vision.scripter.streaming.impl.screen.main.state.YOLO_CLASS
import com.vision.scripter.ui.customClickable

@Composable
fun BaseMenuIcons(
    paramKeys: Set<String>,
    onCvModeClick: () -> Unit,
    onTextClick: () -> Unit,
    onKeyboardClick: () -> Unit,
) {
    val cvKey = paramKeys.firstOrNull { it == TEMPLATE || it == YOLO_CLASS }.orEmpty()
    Icon(
        modifier = Modifier
            .size(32.dp)
            .customClickable(onClick = onCvModeClick),
        imageVector = detectionIcon(cvKey),
        tint = getTint(paramKeys, TEMPLATE, YOLO_CLASS),
        contentDescription = "",
    )

    Icon(
        modifier = Modifier
            .size(32.dp)
            .customClickable(onClick = onTextClick),
        imageVector = Icons.Filled.TextFields,
        tint = getTint(paramKeys, TEXT),
        contentDescription = "",
    )

    Icon(
        modifier = Modifier
            .size(32.dp)
            .customClickable(onClick = onKeyboardClick),
        imageVector = Icons.Filled.Keyboard,
        tint = getTint(paramKeys, TYPE_TEXT),
        contentDescription = "",
    )
}

@Composable
fun getTint(paramKeys: Set<String>, vararg neededParams: String): Color {
    if (paramKeys.any { it in neededParams }) return Color.Red
    return MaterialTheme.colorScheme.onSurface
}

fun detectionIcon(key: String): ImageVector = when (key) {
    YOLO_CLASS -> Icons.Filled.Camera
    TEMPLATE -> Icons.Filled.Visibility
    else -> Icons.Filled.VisibilityOff
}
