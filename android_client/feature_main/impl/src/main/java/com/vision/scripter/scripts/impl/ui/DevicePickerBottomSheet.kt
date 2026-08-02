package com.vision.scripter.scripts.impl.ui

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.Button
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.ModalBottomSheet
import androidx.compose.material3.RadioButton
import androidx.compose.material3.Text
import androidx.compose.material3.rememberModalBottomSheetState
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.vision.scripter.main.impl.R
import com.vision.scripter.scripts.impl.state.DeviceToPick
import com.vision.scripter.ui.customClickable
import kotlinx.collections.immutable.ImmutableList

@OptIn(ExperimentalMaterial3Api::class)
@Composable
internal fun DevicePickerBottomSheet(
    isLoading: Boolean,
    devices: ImmutableList<DeviceToPick>,
    onSelect: (String) -> Unit,
    onConfirm: () -> Unit,
    onDismiss: () -> Unit,
) {
    val sheetState = rememberModalBottomSheetState(skipPartiallyExpanded = true)

    ModalBottomSheet(
        onDismissRequest = onDismiss,
        sheetState = sheetState,
    ) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .padding(horizontal = 16.dp)
                .padding(bottom = 24.dp),
            verticalArrangement = Arrangement.spacedBy(8.dp),
        ) {
            Text(
                text = stringResource(R.string.select_device_title),
                style = TextStyle(
                    fontSize = 20.sp,
                    fontWeight = FontWeight.Bold,
                ),
            )

            when {
                isLoading -> Box(
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(24.dp),
                ) {
                    CircularProgressIndicator(modifier = Modifier.align(Alignment.Center))
                }

                devices.isEmpty() -> Text(
                    modifier = Modifier.padding(vertical = 16.dp),
                    text = stringResource(R.string.devices_not_found),
                    style = TextStyle(fontSize = 16.sp),
                )

                else -> devices.forEach { device ->
                    DeviceRow(
                        device = device,
                        onSelect = { onSelect(device.serial) },
                    )
                }
            }

            Button(
                modifier = Modifier.fillMaxWidth(),
                enabled = devices.isNotEmpty(),
                onClick = onConfirm,
            ) {
                Text(
                    modifier = Modifier.padding(vertical = 4.dp),
                    text = stringResource(R.string.run),
                    style = TextStyle(
                        fontSize = 18.sp,
                        fontWeight = FontWeight.Bold,
                    ),
                )
            }
        }
    }
}

@Composable
private fun DeviceRow(
    device: DeviceToPick,
    onSelect: () -> Unit,
) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .customClickable(onClick = onSelect)
            .padding(vertical = 4.dp),
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.spacedBy(8.dp),
    ) {
        RadioButton(
            selected = device.selected,
            onClick = null,
        )
        Text(
            text = device.title,
            style = TextStyle(fontSize = 18.sp),
        )
    }
}
