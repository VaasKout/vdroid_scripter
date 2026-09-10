package com.vision.scripter.streaming.impl.blocks.menu.ui.menu

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.automirrored.filled.ArrowForward
import androidx.compose.material.icons.automirrored.filled.Logout
import androidx.compose.material.icons.filled.Add
import androidx.compose.material.icons.filled.Keyboard
import androidx.compose.material.icons.filled.Search
import androidx.compose.material.icons.filled.Visibility
import androidx.compose.material.icons.filled.VisibilityOff
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
fun UsualMenu(
    modifier: Modifier = Modifier,
    menuType: MenuType.Usual,
    uiStateHolder: MenuUiStateHolder,
) {
    if (menuType.expanded) {
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
                    .customClickable(onClick = uiStateHolder::onAddClicked),
                imageVector = Icons.Filled.Add,
                tint = MaterialTheme.colorScheme.onSurface,
                contentDescription = ""
            )

            ScanIcon(
                menuType = menuType,
                onClick = uiStateHolder::onScanClicked,
            )

            RectanglesIcon(
                menuType = menuType,
                onClick = uiStateHolder::onRectanglesClicked,
            )

            Icon(
                modifier = Modifier
                    .size(32.dp)
                    .customClickable(onClick = uiStateHolder::onKeyboardClicked),
                imageVector = Icons.Filled.Keyboard,
                tint = MaterialTheme.colorScheme.onSurface,
                contentDescription = "",
            )

            Icon(
                modifier = Modifier
                    .size(32.dp)
                    .customClickable(onClick = uiStateHolder::onExitClicked),
                imageVector = Icons.AutoMirrored.Filled.Logout,
                tint = MaterialTheme.colorScheme.onSurface,
                contentDescription = ""
            )

            Icon(
                modifier = Modifier
                    .size(32.dp)
                    .customClickable(onClick = uiStateHolder::onExpandClicked),
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
            .customClickable(onClick = uiStateHolder::onExpandClicked),
        imageVector = Icons.AutoMirrored.Filled.ArrowBack,
        tint = MaterialTheme.colorScheme.onSurface,
        contentDescription = ""
    )
}

@Composable
private fun RectanglesIcon(
    menuType: MenuType.Usual,
    onClick: () -> Unit,
) {
    if (menuType.rectsAreLoading) {
        CircularProgressIndicator(modifier = Modifier.size(32.dp))
        return
    }
    Icon(
        modifier = Modifier
            .size(32.dp)
            .customClickable(onClick = onClick),
        imageVector = if (menuType.rectsShown) Icons.Filled.Visibility
        else Icons.Filled.VisibilityOff,
        tint = if (menuType.rectsShown) Color.Red
        else MaterialTheme.colorScheme.onSurface,
        contentDescription = "",
    )
}

@Composable
private fun ScanIcon(
    menuType: MenuType.Usual,
    onClick: () -> Unit,
) {
    if (menuType.scanning) {
        CircularProgressIndicator(modifier = Modifier.size(32.dp))
        return
    }
    Icon(
        modifier = Modifier
            .size(32.dp)
            .customClickable(onClick = onClick),
        imageVector = Icons.Filled.Search,
        tint = if (menuType.scanShown) Color.Red
        else MaterialTheme.colorScheme.onSurface,
        contentDescription = "",
    )
}

@Preview
@Composable
private fun UsualMenuPreview() {
    UsualMenu(
        menuType = MenuType.Usual(),
        uiStateHolder = MenuPreviewUiStateHolder(usualMenuPreviewUiState),
    )
}

@Preview
@Composable
private fun UsualExpandedMenuPreview() {
    UsualMenu(
        menuType = MenuType.Usual(expanded = true),
        uiStateHolder = MenuPreviewUiStateHolder(usualMenuPreviewUiState),
    )
}

@Preview
@Composable
private fun UsualExpandedHighlightedPreview() {
    UsualMenu(
        menuType = MenuType.Usual(
            expanded = true,
            rectsShown = true,
        ),
        uiStateHolder = MenuPreviewUiStateHolder(usualMenuPreviewUiState),
    )
}
