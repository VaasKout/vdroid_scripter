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
import androidx.compose.ui.unit.dp
import com.vision.scripter.data.api.models.CLASS_IS_VISIBLE
import com.vision.scripter.data.api.models.EVENT_ON_CLASS
import com.vision.scripter.data.api.models.EVENT_ON_TEMPLATE
import com.vision.scripter.data.api.models.EVENT_ON_TEXT
import com.vision.scripter.data.api.models.TEMPLATE_IS_VISIBLE
import com.vision.scripter.data.api.models.TEXT_IS_VISIBLE
import com.vision.scripter.data.api.models.TYPE_TEXT
import com.vision.scripter.streaming.impl.screen.main.state.classFlag
import com.vision.scripter.streaming.impl.screen.main.state.hasFlag
import com.vision.scripter.streaming.impl.screen.main.state.templateFlag
import com.vision.scripter.ui.customClickable

@Composable
fun BaseMenuIcons(
    flags: Int,
    onCvModeClick: () -> Unit,
    onTextClick: () -> Unit,
    onKeyboardClick: () -> Unit,
) {
    Icon(
        modifier = Modifier
            .size(32.dp)
            .customClickable(onClick = onCvModeClick),
        imageVector = when {
            flags.classFlag() != 0 -> Icons.Filled.Camera
            flags.templateFlag() != 0 -> Icons.Filled.Visibility
            else -> Icons.Filled.VisibilityOff
        },
        tint = detectionTint(flags),
        contentDescription = "",
    )

    Icon(
        modifier = Modifier
            .size(32.dp)
            .customClickable(onClick = onTextClick),
        imageVector = Icons.Filled.TextFields,
        tint = textTint(flags),
        contentDescription = "",
    )

    Icon(
        modifier = Modifier
            .size(32.dp)
            .customClickable(onClick = onKeyboardClick),
        imageVector = Icons.Filled.Keyboard,
        tint = if (flags.hasFlag(TYPE_TEXT)) Color.Red
        else MaterialTheme.colorScheme.onSurface,
        contentDescription = "",
    )
}

@Composable
fun detectionTint(flags: Int): Color = when {
    flags.hasFlag(TEMPLATE_IS_VISIBLE) || flags.hasFlag(CLASS_IS_VISIBLE) -> Color.Blue
    flags.hasFlag(EVENT_ON_TEMPLATE) || flags.hasFlag(EVENT_ON_CLASS) -> Color.Red
    else -> MaterialTheme.colorScheme.onSurface
}

@Composable
fun textTint(flags: Int): Color = when {
    flags.hasFlag(TEXT_IS_VISIBLE) -> Color.Blue
    flags.hasFlag(EVENT_ON_TEXT) -> Color.Red
    else -> MaterialTheme.colorScheme.onSurface
}
