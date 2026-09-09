package com.vision.scripter.libraryitems.ui.items

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.ModalBottomSheet
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
import com.vision.scripter.ui.customClickable

@OptIn(ExperimentalMaterial3Api::class)
@Composable
internal fun DevicePickerSheet(
    picker: UiDevicePicker,
    onDeviceChosen: (String) -> Unit,
    onDismiss: () -> Unit,
) {
    ModalBottomSheet(
        onDismissRequest = onDismiss,
        sheetState = rememberModalBottomSheetState(),
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
                    onClick = { onDeviceChosen(device.serial) },
                )
            }
        }
    }
}

@Composable
private fun PickerDeviceRow(
    device: UiPickerDevice,
    onClick: () -> Unit,
) {
    val rowModifier = if (device.busy) Modifier
    else Modifier.customClickable(onClick = onClick)
    Column(
        modifier = Modifier
            .fillMaxWidth()
            .then(rowModifier)
            .padding(12.dp),
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
        if (device.busy) {
            Text(
                text = stringResource(R.string.busy_device, device.statusText),
                style = TextStyle(
                    color = Color.Red,
                    fontSize = 12.sp,
                    fontWeight = FontWeight.Normal,
                )
            )
            return
        }
        if (device.lastUsed) {
            Text(
                text = stringResource(R.string.last_used),
                style = TextStyle(
                    color = Color.Gray,
                    fontSize = 12.sp,
                    fontWeight = FontWeight.Normal,
                )
            )
        }
    }
}
