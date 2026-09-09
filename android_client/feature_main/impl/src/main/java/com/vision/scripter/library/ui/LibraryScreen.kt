package com.vision.scripter.library.ui

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Gesture
import androidx.compose.material.icons.filled.Image
import androidx.compose.material.icons.filled.Route
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.tooling.preview.Preview
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.vision.scripter.library.state.LibraryType
import com.vision.scripter.library.state.LibraryUiState
import com.vision.scripter.library.state.LibraryUiStateHolder
import com.vision.scripter.main.impl.R
import com.vision.scripter.ui.CustomPullToRefresh
import com.vision.scripter.ui.customClickable

@Composable
internal fun LibraryScreen(
    uiStateHolder: LibraryUiStateHolder,
    paddingValues: PaddingValues,
) {
    val state = uiStateHolder.uiStateFlow.collectAsStateWithLifecycle(
        initialValue = LibraryUiState(),
    ).value

    LaunchedEffect(Unit) {
        uiStateHolder.onLoadData(onStart = true)
    }

    if (state.isLoading) {
        Box(
            modifier = Modifier
                .fillMaxSize()
                .padding(paddingValues),
        ) {
            CircularProgressIndicator(modifier = Modifier.align(Alignment.Center))
        }
        return
    }

    CustomPullToRefresh(
        modifier = Modifier
            .fillMaxSize()
            .padding(paddingValues),
        isRefreshing = state.isRefreshing,
        onRefresh = {
            uiStateHolder.onLoadData(onStart = false)
        }
    ) {
        Column(
            modifier = Modifier
                .fillMaxSize()
                .verticalScroll(rememberScrollState())
                .padding(16.dp),
            verticalArrangement = Arrangement.spacedBy(8.dp),
        ) {
            LibraryCard(
                title = stringResource(R.string.images),
                count = state.imagesCount,
                icon = Icons.Filled.Image,
                onClick = { uiStateHolder.onCardClicked(LibraryType.IMAGES) },
            )
            LibraryCard(
                title = stringResource(R.string.actions),
                count = state.actionsCount,
                icon = Icons.Filled.Gesture,
                onClick = { uiStateHolder.onCardClicked(LibraryType.ACTIONS) },
            )
            LibraryCard(
                title = stringResource(R.string.routes),
                count = state.routesCount,
                icon = Icons.Filled.Route,
                onClick = { uiStateHolder.onCardClicked(LibraryType.ROUTES) },
            )
        }
    }
}

@Composable
private fun LibraryCard(
    modifier: Modifier = Modifier,
    title: String,
    count: Int,
    icon: ImageVector,
    onClick: () -> Unit,
) {
    Row(
        modifier = modifier
            .fillMaxWidth()
            .background(color = Color.White, shape = RoundedCornerShape(16.dp))
            .customClickable(onClick = onClick)
            .padding(16.dp),
        horizontalArrangement = Arrangement.spacedBy(16.dp),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Icon(
            modifier = Modifier.size(40.dp),
            imageVector = icon,
            tint = MaterialTheme.colorScheme.primary,
            contentDescription = title,
        )
        Column(modifier = Modifier.weight(1f)) {
            Text(
                text = title,
                style = TextStyle(
                    color = Color.Black,
                    fontSize = 24.sp,
                    fontWeight = FontWeight.Bold,
                )
            )
            Text(
                text = stringResource(R.string.items_count, count),
                style = TextStyle(
                    color = Color.Gray,
                    fontSize = 14.sp,
                    fontWeight = FontWeight.Normal,
                )
            )
        }
    }
}

@Preview
@Composable
private fun LibraryScreenPreview() {
    LibraryScreen(
        uiStateHolder = LibraryUiStateHolderPreview(libraryUiStatePreview),
        paddingValues = PaddingValues(0.dp),
    )
}
