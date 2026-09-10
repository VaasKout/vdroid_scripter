package com.vision.scripter.libraryitems.ui.items

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.selection.selectable
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.ModalBottomSheet
import androidx.compose.material3.RadioButton
import androidx.compose.material3.Text
import androidx.compose.material3.rememberModalBottomSheetState
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.vision.scripter.libraryitems.ui.UiDevicePicker
import com.vision.scripter.libraryitems.ui.UiPickerDevice
import com.vision.scripter.main.impl.R
import com.vision.scripter.ui.CustomButton

@OptIn(ExperimentalMaterial3Api::class)
@Composable
internal fun DevicePickerSheet(
    picker: UiDevicePicker,
    onDeviceSelected: (String) -> Unit,
    onPlayClick: () -> Unit,
    onDismiss: () -> Unit,
) {
    ModalBottomSheet(
        onDismissRequest = onDismiss,
        sheetState = rememberModalBottomSheetState(skipPartiallyExpanded = true),
    ) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .padding(horizontal = 16.dp)
                .padding(bottom = 32.dp),
            verticalArrangement = Arrangement.spacedBy(8.dp),
        ) {
            Text(
                text = stringResource(R.string.pick_device),
                style = TextStyle(
                    color = Color.Black,
                    fontSize = 18.sp,
                    fontWeight = FontWeight.Bold,
                )
            )
            if (picker.isLoading) {
                CircularProgressIndicator(
                    modifier = Modifier.align(Alignment.CenterHorizontally),
                )
                return@Column
            }
            if (picker.devices.isEmpty()) {
                Text(
                    text = stringResource(R.string.devices_not_found),
                    style = TextStyle(
                        color = Color.Gray,
                        fontSize = 16.sp,
                        fontWeight = FontWeight.Normal,
                    )
                )
                return@Column
            }
            picker.devices.forEach { device ->
                PickerDeviceRow(
                    device = device,
                    onClick = { onDeviceSelected(device.serial) },
                )
            }
            CustomButton(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(top = 8.dp),
                text = stringResource(R.string.play),
                enabled = picker.canPlay,
                onClick = onPlayClick,
            )
        }
    }
}

@Composable
private fun PickerDeviceRow(
    device: UiPickerDevice,
    onClick: () -> Unit,
) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .selectable(
                selected = device.selected,
                enabled = !device.busy,
                onClick = onClick,
            )
            .padding(vertical = 4.dp),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        RadioButton(
            selected = device.selected,
            enabled = !device.busy,
            onClick = onClick,
        )
        Column(
            verticalArrangement = Arrangement.spacedBy(2.dp),
        ) {
            Text(
                text = device.label,
                style = TextStyle(
                    color = if (device.busy) Color.Gray else Color.Black,
                    fontSize = 18.sp,
                    fontWeight = FontWeight.Normal,
                )
            )
            if (!device.busy) return@Column
            Text(
                text = stringResource(R.string.busy_device, device.statusText),
                style = TextStyle(
                    color = Color.Red,
                    fontSize = 12.sp,
                    fontWeight = FontWeight.Normal,
                )
            )
        }
    }
}
