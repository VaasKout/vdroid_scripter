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
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.tooling.preview.Preview
import androidx.compose.ui.unit.dp
import com.vision.scripter.streaming.impl.blocks.menu.state.MenuType
import com.vision.scripter.streaming.impl.blocks.menu.ui.MenuPreviewUiStateHolder
import com.vision.scripter.streaming.impl.blocks.menu.ui.MenuUiStateHolder
import com.vision.scripter.streaming.impl.blocks.menu.ui.usualMenuPreviewUiState
import com.vision.scripter.ui.customClickable

@Composable
fun CustomActionMenu(
    modifier: Modifier = Modifier,
    menuType: MenuType.CustomAction,
    uiStateHolder: MenuUiStateHolder,
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
                    onClick = uiStateHolder::onRecordingClicked,
                ),
            imageVector = Icons.Filled.RadioButtonChecked,
            tint =
                if (menuType.recording) Color.Red
                else MaterialTheme.colorScheme.onSurface,
            contentDescription = ""
        )

        Icon(
            modifier = Modifier
                .size(32.dp)
                .customClickable(onClick = uiStateHolder::onSaveClicked),
            imageVector = Icons.Filled.Check,
            tint = Color.Green,
            contentDescription = ""
        )

        Icon(
            modifier = Modifier
                .size(32.dp)
                .customClickable(onClick = uiStateHolder::onCancelClicked),
            imageVector = Icons.Filled.Close,
            tint = MaterialTheme.colorScheme.onSurface,
            contentDescription = ""
        )
    }
}

@Preview
@Composable
private fun CustomActionMenuPreview() {
    CustomActionMenu(
        menuType = MenuType.CustomAction(),
        uiStateHolder = MenuPreviewUiStateHolder(usualMenuPreviewUiState),
    )
}

@Preview
@Composable
private fun CustomActionMenuRecordingPreview() {
    CustomActionMenu(
        menuType = MenuType.CustomAction(recording = true),
        uiStateHolder = MenuPreviewUiStateHolder(usualMenuPreviewUiState),
    )
}
