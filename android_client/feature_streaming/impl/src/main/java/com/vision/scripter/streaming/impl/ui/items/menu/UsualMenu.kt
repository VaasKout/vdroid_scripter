package com.vision.scripter.streaming.impl.ui.items.menu

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.automirrored.filled.ArrowForward
import androidx.compose.material.icons.filled.Add
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
import com.vision.scripter.streaming.impl.state.CVMode
import com.vision.scripter.streaming.impl.state.MenuState
import com.vision.scripter.streaming.impl.state.withFlag
import com.vision.scripter.ui.customClickable

@Composable
fun UsualMenu(
    modifier: Modifier = Modifier,
    menuState: MenuState.Usual,
    onScriptModeClick: () -> Unit,
    onKeyboardClick: () -> Unit,
    onCvModeClick: () -> Unit,
    onTextModeClick: () -> Unit,
    onExpandClick: () -> Unit,
) {
    if (menuState.expanded) {
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
                        onClick = onScriptModeClick,
                    ),
                imageVector = Icons.Filled.Add,
                tint = MaterialTheme.colorScheme.onSurface,
                contentDescription = ""
            )

            BaseMenuIcons(
                flags = menuState.toFlags(),
                onCvModeClick = onCvModeClick,
                onTextClick = onTextModeClick,
                keyboardHighlighted = false,
                onKeyboardClick = onKeyboardClick,
            )

            Icon(
                modifier = Modifier
                    .size(32.dp)
                    .customClickable(onClick = onExpandClick),
                imageVector = Icons.AutoMirrored.Filled.ArrowForward,
                tint = MaterialTheme.colorScheme.onSurface,
                contentDescription = ""
            )
        }

        return
    }

    Icon(
        modifier = modifier
            .size(32.dp)
            .background(
                color = Color.White,
                shape = RoundedCornerShape(topStart = 8.dp, bottomStart = 8.dp),
            )
            .customClickable(
                onClick = onExpandClick,
            ),
        imageVector = Icons.AutoMirrored.Filled.ArrowBack,
        tint = MaterialTheme.colorScheme.onSurface,
        contentDescription = ""
    )
}

private fun MenuState.Usual.toFlags(): Int =
    0.withFlag(EVENT_ON_TEMPLATE, cvMode == CVMode.CV_RECTS)
        .withFlag(EVENT_ON_TEXT, textHighlighted)

@Preview
@Composable
private fun UsualMenuPreview() {
    UsualMenu(
        menuState = MenuState.Usual(),
        onScriptModeClick = {},
        onKeyboardClick = {},
        onCvModeClick = {},
        onTextModeClick = {},
        onExpandClick = {},
    )
}

@Preview
@Composable
private fun UsualExpandedMenuPreview() {
    UsualMenu(
        menuState = MenuState.Usual(expanded = true),
        onScriptModeClick = {},
        onKeyboardClick = {},
        onCvModeClick = {},
        onTextModeClick = {},
        onExpandClick = {},
    )
}

@Preview
@Composable
private fun UsualExpandedHighlightedTextPreview() {
    UsualMenu(
        menuState = MenuState.Usual(
            expanded = true,
            textHighlighted = true,
            cvMode = CVMode.CV_RECTS,
        ),
        onScriptModeClick = {},
        onKeyboardClick = {},
        onCvModeClick = {},
        onTextModeClick = {},
        onExpandClick = {},
    )
}
