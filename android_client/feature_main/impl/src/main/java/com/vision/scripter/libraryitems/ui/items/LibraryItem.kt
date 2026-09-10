package com.vision.scripter.libraryitems.ui.items

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Delete
import androidx.compose.material.icons.filled.PlayArrow
import androidx.compose.material3.Icon
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.tooling.preview.Preview
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.vision.scripter.libraryitems.ui.UiLibraryItem
import com.vision.scripter.libraryitems.ui.UiRun
import com.vision.scripter.main.impl.R
import com.vision.scripter.ui.CustomColors
import com.vision.scripter.ui.customClickable

@Composable
internal fun LibraryItem(
    modifier: Modifier = Modifier,
    item: UiLibraryItem,
    onPlayClick: (String) -> Unit,
    onDeleteClick: (String) -> Unit,
    onRunDismiss: (String) -> Unit,
) {
    Column(
        modifier = modifier
            .fillMaxWidth()
            .background(color = Color.White, shape = RoundedCornerShape(16.dp))
            .padding(16.dp),
        verticalArrangement = Arrangement.spacedBy(4.dp),
    ) {
        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.SpaceBetween,
            verticalAlignment = Alignment.CenterVertically,
        ) {
            Text(
                modifier = Modifier.weight(1f),
                text = item.name,
                style = TextStyle(
                    color = Color.Black,
                    fontSize = 24.sp,
                    fontWeight = FontWeight.Bold,
                )
            )
            Spacer(modifier = Modifier.width(4.dp))
            if (item.canPlay) {
                Icon(
                    modifier = Modifier
                        .size(32.dp)
                        .customClickable(onClick = { onPlayClick(item.name) }),
                    imageVector = Icons.Filled.PlayArrow,
                    tint = CustomColors.LightGreen,
                    contentDescription = stringResource(R.string.play),
                )
                Spacer(modifier = Modifier.width(4.dp))
            }
            Icon(
                modifier = Modifier
                    .size(32.dp)
                    .customClickable(onClick = { onDeleteClick(item.name) }),
                imageVector = Icons.Filled.Delete,
                tint = Color.Red,
                contentDescription = stringResource(R.string.delete),
            )
        }
        val run = item.run ?: return@Column
        RunStatus(
            run = run,
            onDismiss = { onRunDismiss(item.name) },
        )
    }
}

@Composable
private fun RunStatus(
    run: UiRun,
    onDismiss: () -> Unit,
) {
    val statusModifier = if (run.isError) Modifier.customClickable(onClick = onDismiss)
    else Modifier
    Text(
        modifier = statusModifier.fillMaxWidth(),
        text = run.statusText,
        style = TextStyle(
            color = if (run.isError) Color.Red else Color.Gray,
            fontSize = 14.sp,
            fontWeight = FontWeight.Normal,
        )
    )
    Text(
        text = stringResource(R.string.running_on, run.deviceLabel),
        style = TextStyle(
            color = Color.Gray,
            fontSize = 12.sp,
            fontWeight = FontWeight.Normal,
        )
    )
}

@Preview
@Composable
private fun LibraryItemPreview() {
    LibraryItem(
        modifier = Modifier.fillMaxWidth(),
        item = UiLibraryItem(name = "shop_catalog_swipe_1", canPlay = true),
        onPlayClick = {},
        onDeleteClick = {},
        onRunDismiss = {},
    )
}

@Preview
@Composable
private fun LibraryItemRunningPreview() {
    LibraryItem(
        modifier = Modifier.fillMaxWidth(),
        item = UiLibraryItem(
            name = "shop_checkout",
            canPlay = true,
            run = UiRun(
                statusText = "running step 2: tap on text Checkout",
                deviceLabel = "Pixel 6 (emulator-5554)",
                isError = false,
            ),
        ),
        onPlayClick = {},
        onDeleteClick = {},
        onRunDismiss = {},
    )
}
