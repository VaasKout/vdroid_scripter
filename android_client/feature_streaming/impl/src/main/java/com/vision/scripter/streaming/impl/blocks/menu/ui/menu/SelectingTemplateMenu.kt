package com.vision.scripter.streaming.impl.blocks.menu.ui.menu

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.Check
import androidx.compose.material.icons.filled.Refresh
import androidx.compose.material3.CircularProgressIndicator
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
fun SelectingTemplateMenu(
    modifier: Modifier = Modifier,
    menuType: MenuType.SelectingCV,
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
        RefreshIcon(
            menuType = menuType,
            onClick = uiStateHolder::onRefreshRectanglesClicked,
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
            imageVector = Icons.AutoMirrored.Filled.ArrowBack,
            tint = MaterialTheme.colorScheme.onSurface,
            contentDescription = ""
        )
    }
}

@Composable
private fun RefreshIcon(
    menuType: MenuType.SelectingCV,
    onClick: () -> Unit,
) {
    if (menuType.rectsLoading) {
        CircularProgressIndicator(modifier = Modifier.size(32.dp))
        return
    }
    Icon(
        modifier = Modifier
            .size(32.dp)
            .customClickable(onClick = onClick),
        imageVector = Icons.Filled.Refresh,
        tint = Color.Red,
        contentDescription = ""
    )
}

@Preview
@Composable
fun SelectingTemplateMenuPreview() {
    SelectingTemplateMenu(
        menuType = MenuType.SelectingCV(),
        uiStateHolder = MenuPreviewUiStateHolder(usualMenuPreviewUiState),
    )
}

@Preview
@Composable
fun SelectingTemplateMenuLoadingPreview() {
    SelectingTemplateMenu(
        menuType = MenuType.SelectingCV(rectsLoading = true),
        uiStateHolder = MenuPreviewUiStateHolder(usualMenuPreviewUiState),
    )
}
