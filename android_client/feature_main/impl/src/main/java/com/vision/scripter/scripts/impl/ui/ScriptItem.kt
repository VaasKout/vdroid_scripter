package com.vision.scripter.scripts.impl.ui

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Delete
import androidx.compose.material.icons.filled.Edit
import androidx.compose.material.icons.filled.PlayArrow
import androidx.compose.material3.Icon
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.tooling.preview.Preview
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.vision.scripter.ui.customClickable

@Composable
fun ScriptItem(
    modifier: Modifier = Modifier,
    name: String,
    onPlayClick: (String) -> Unit,
    onEditClick: (String) -> Unit,
    onDeleteClick: (String) -> Unit,
) {
    Row(
        modifier = modifier
            .fillMaxWidth()
            .background(
                color = Color.White,
                shape = RoundedCornerShape(16.dp)
            )
            .padding(16.dp),
        horizontalArrangement = Arrangement.SpaceBetween,
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Text(
            modifier = Modifier.weight(1f),
            text = name,
            color = Color.Black,
            style = TextStyle(
                color = Color.Black,
                fontSize = 24.sp,
                fontWeight = FontWeight.Bold,
            )
        )
        Spacer(modifier = Modifier.width(4.dp))
        Row(
            horizontalArrangement = Arrangement.spacedBy(4.dp),
        ) {
            Icon(
                modifier = Modifier
                    .size(32.dp)
                    .customClickable(onClick = { onEditClick(name) }),
                imageVector = Icons.Filled.Edit,
                tint = Color.Gray,
                contentDescription = "",
            )
            Icon(
                modifier = Modifier
                    .size(32.dp)
                    .customClickable(onClick = { onDeleteClick(name) }),
                imageVector = Icons.Filled.Delete,
                tint = Color.Red,
                contentDescription = "",
            )
            Icon(
                modifier = Modifier
                    .size(32.dp)
                    .customClickable(
                        onClick = {
                            onPlayClick(name)
                        },
                    ),
                imageVector = Icons.Filled.PlayArrow,
                tint = Color.Green,
                contentDescription = "",
            )
        }
    }
}

@Composable
fun LocationItem(
    modifier: Modifier = Modifier,
    name: String,
    onClick: (String) -> Unit,
    onDeleteClick: (String) -> Unit,
) {
    Row(
        modifier = modifier
            .fillMaxWidth()
            .background(color = Color.White, shape = RoundedCornerShape(16.dp))
            .customClickable(onClick = { onClick(name) })
            .padding(16.dp),
        horizontalArrangement = Arrangement.SpaceBetween,
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Text(
            modifier = Modifier.weight(1f),
            text = name,
            color = Color.Black,
            style = TextStyle(
                color = Color.Black,
                fontSize = 24.sp,
                fontWeight = FontWeight.Bold,
            )
        )
        Spacer(modifier = Modifier.width(4.dp))
        Icon(
            modifier = Modifier
                .size(32.dp)
                .customClickable(onClick = { onDeleteClick(name) }),
            imageVector = Icons.Filled.Delete,
            tint = Color.Red,
            contentDescription = "",
        )
    }
}

@Preview
@Composable
private fun ScriptItemPreview() {
    ScriptItem(
        modifier = Modifier.fillMaxWidth(),
        name = "test_1",
        onPlayClick = {},
        onEditClick = {},
        onDeleteClick = {},
    )
}

@Preview
@Composable
private fun ScriptItemLongNamePreview() {
    ScriptItem(
        modifier = Modifier.fillMaxWidth(),
        name = "very_long_script_name_that_wraps_to_multiple_lines",
        onPlayClick = {},
        onEditClick = {},
        onDeleteClick = {},
    )
}

@Preview
@Composable
private fun LocationItemPreview() {
    LocationItem(
        modifier = Modifier.fillMaxWidth(),
        name = "main_screen",
        onClick = {},
        onDeleteClick = {},
    )
}

@Preview
@Composable
private fun LocationLongItemPreview() {
    LocationItem(
        modifier = Modifier.fillMaxWidth(),
        name = "screen_with_a_very_long_name",
        onClick = {},
        onDeleteClick = {},
    )
}