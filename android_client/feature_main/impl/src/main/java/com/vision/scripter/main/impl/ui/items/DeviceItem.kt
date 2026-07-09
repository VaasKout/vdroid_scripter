package com.vision.scripter.main.impl.ui.items

import androidx.compose.foundation.Image
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.PhoneIphone
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.SpanStyle
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.text.buildAnnotatedString
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.withStyle
import androidx.compose.ui.tooling.preview.Preview
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.vision.scripter.main.impl.R
import com.vision.scripter.main.impl.ui.UiDevice
import com.vision.scripter.ui.CustomColors
import com.vision.scripter.ui.customClickable
import kotlinx.collections.immutable.persistentMapOf

@Composable
internal fun DeviceItem(
    modifier: Modifier = Modifier,
    uiDevice: UiDevice,
    onStreamingClick: () -> Unit,
) {
    Box(
        modifier = modifier
            .fillMaxWidth()
            .background(color = CustomColors.LightGreen, shape = RoundedCornerShape(16.dp))
            .padding(16.dp),
    ) {
        Column(
            modifier = Modifier.align(Alignment.Center),
            verticalArrangement = Arrangement.spacedBy(4.dp)
        ) {
            Row(
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically,
            ) {
                if (uiDevice.preview != null) {
                    Image(
                        modifier = Modifier.size(128.dp),
                        contentScale = ContentScale.Fit,
                        bitmap = uiDevice.preview,
                        contentDescription = ""
                    )
                } else {
                    Image(
                        modifier = Modifier.size(128.dp),
                        contentScale = ContentScale.Fit,
                        imageVector = Icons.Default.PhoneIphone,
                        contentDescription = ""
                    )
                }
                Column {
                    uiDevice.deviceParams.forEach { (key, value) ->
                        val keyWord = stringResource(key)
                        Text(
                            text = buildAnnotatedString {
                                withStyle(
                                    style = SpanStyle(
                                        color = Color.Gray,
                                        fontSize = 24.sp,
                                        fontWeight = FontWeight.Normal
                                    ),
                                ) {
                                    append(keyWord)
                                    append(":")
                                }
                                append(" ")
                                withStyle(
                                    style = SpanStyle(
                                        color = Color.Black,
                                        fontSize = 24.sp,
                                        fontWeight = FontWeight.Normal
                                    ),
                                ) {
                                    append(value)
                                }
                            }
                        )
                    }
                }
            }
            DeviceItemButton(
                modifier = Modifier.padding(top = 8.dp),
                text = stringResource(R.string.streaming),
                onClick = onStreamingClick,
            )
        }
    }
}

@Composable
private fun DeviceItemButton(
    modifier: Modifier = Modifier,
    text: String,
    onClick: () -> Unit,
) {
    Box(
        modifier = modifier
            .fillMaxWidth()
            .background(color = Color.White, shape = RoundedCornerShape(16.dp))
            .customClickable(onClick = onClick)
            .padding(16.dp),
    ) {
        Text(
            modifier = Modifier.align(Alignment.Center),
            text = text,
            color = Color.Black,
            style = TextStyle(
                color = Color.Black,
                fontSize = 24.sp,
                fontWeight = FontWeight.Bold
            )
        )
    }
}

@Preview
@Composable
private fun DeviceItemPreview() {
    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(color = MaterialTheme.colorScheme.background)
    ) {
        DeviceItem(
            modifier = Modifier
                .fillMaxWidth()
                .padding(horizontal = 8.dp)
                .padding(top = 8.dp)
                .align(Alignment.TopStart),
            uiDevice = UiDevice(
                serial = "c88893qa",
                deviceParams = persistentMapOf(
                    R.string.serial_key to "c9858321",
                    R.string.model_key to "Samsung",
                ),
            ),
            onStreamingClick = {},
        )
    }
}
