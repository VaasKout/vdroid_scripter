package com.vision.scripter.devices.ui

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.tooling.preview.Preview
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.vision.scripter.main.impl.R
import com.vision.scripter.devices.ui.items.DeviceItem
import com.vision.scripter.ui.CustomPullToRefresh

@Composable
internal fun DevicesScreen(
    uiStateHolder: DevicesUiStateHolder,
    paddingValues: PaddingValues,
) {
    val state = uiStateHolder.uiStateFlow.collectAsStateWithLifecycle(
        initialValue = DevicesUiState(),
    ).value

    LaunchedEffect(Unit) {
        uiStateHolder.onLoadData(onStart = true)
    }

    CustomPullToRefresh(
        modifier = Modifier
            .fillMaxSize()
            .padding(paddingValues),
        isRefreshing = state.isRefreshing,
        onRefresh = { uiStateHolder.onLoadData(onStart = false) }
    ) {
        if (state.isLoading && state.devices.isEmpty()) {
            CircularProgressIndicator(modifier = Modifier.align(Alignment.Center))
            return@CustomPullToRefresh
        }

        if (state.devices.isEmpty()) {
            Text(
                modifier = Modifier.align(Alignment.Center),
                text = stringResource(R.string.devices_not_found),
                style = TextStyle(
                    fontSize = 18.sp,
                    fontWeight = FontWeight.Bold
                )
            )
            return@CustomPullToRefresh
        }

        LazyColumn(
            modifier = Modifier.fillMaxSize(),
            contentPadding = PaddingValues(8.dp),
            verticalArrangement = Arrangement.spacedBy(8.dp)
        ) {
            items(items = state.devices, key = { it.serial }) {
                DeviceItem(
                    modifier = Modifier.fillMaxWidth(),
                    uiDevice = it,
                    onStreamingClick = {
                        uiStateHolder.uiCommandsFlow.tryEmit(
                            DevicesUiCommand.NavigateToStreaming(it.serial)
                        )
                    },
                )
            }
        }
    }
}

@Preview
@Composable
private fun DevicesLoadingScreenPreview() {
    DevicesScreen(
        uiStateHolder = DevicesScreenUiStateHolderPreview(devicesUiStateLoadingPreview),
        paddingValues = PaddingValues(0.dp),
    )
}

@Preview
@Composable
private fun DevicesNoDataScreenPreview() {
    DevicesScreen(
        uiStateHolder = DevicesScreenUiStateHolderPreview(devicesUiStateNoDataPreview),
        paddingValues = PaddingValues(0.dp),
    )
}

@Preview
@Composable
private fun DevicesWithDataScreenPreview() {
    DevicesScreen(
        uiStateHolder = DevicesScreenUiStateHolderPreview(devicesUiStateWithDataPreview),
        paddingValues = PaddingValues(0.dp),
    )
}