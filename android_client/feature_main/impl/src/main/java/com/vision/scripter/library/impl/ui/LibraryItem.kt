package com.vision.scripter.library.impl.ui

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
fun LibraryItem(
    modifier: Modifier = Modifier,
    name: String,
    onDeleteClick: (String) -> Unit,
) {
    Row(
        modifier = modifier
            .fillMaxWidth()
            .background(color = Color.White, shape = RoundedCornerShape(16.dp))
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
private fun LibraryItemPreview() {
    LibraryItem(
        modifier = Modifier.fillMaxWidth(),
        name = "x5_catalog_cart_icon",
        onDeleteClick = {},
    )
}

@Preview
@Composable
private fun LibraryItemLongNamePreview() {
    LibraryItem(
        modifier = Modifier.fillMaxWidth(),
        name = "very_long_custom_action_name_that_wraps_to_lines",
        onDeleteClick = {},
    )
}
