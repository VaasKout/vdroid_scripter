package com.vision.scripter.editscript.impl.ui

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Delete
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
import com.vision.scripter.editscript.impl.R
import com.vision.scripter.ui.customClickable

@Composable
internal fun ParamCard(
    modifier: Modifier = Modifier,
    param: ParamUiData,
    onDeleteClick: (Int) -> Unit,
) {
    Box(
        modifier = modifier
            .fillMaxWidth()
            .background(color = Color.White, shape = RoundedCornerShape(16.dp))
            .padding(16.dp),
    ) {
        Column(
            modifier = Modifier
                .align(Alignment.CenterStart)
                .padding(end = 40.dp),
        ) {
            Text(
                text = param.title,
                style = TextStyle(
                    color = Color.Black,
                    fontSize = 18.sp,
                    fontWeight = FontWeight.Bold,
                ),
            )
            if (param.locale.isNotEmpty()) {
                Text(
                    text = param.locale,
                    style = TextStyle(
                        color = Color.DarkGray,
                        fontSize = 14.sp,
                    ),
                )
            }
        }

        Icon(
            modifier = Modifier
                .align(Alignment.CenterEnd)
                .size(32.dp)
                .customClickable(onClick = { onDeleteClick(param.id) }),
            imageVector = Icons.Filled.Delete,
            tint = Color.Red,
            contentDescription = "",
        )
    }
}

@Composable
internal fun EventsCard(
    modifier: Modifier = Modifier,
    count: Int,
    onDeleteClick: () -> Unit,
) {
    Box(
        modifier = modifier
            .fillMaxWidth()
            .background(color = Color.White, shape = RoundedCornerShape(16.dp))
            .padding(16.dp),
    ) {
        Text(
            modifier = Modifier
                .align(Alignment.CenterStart)
                .padding(end = 40.dp),
            text = stringResource(R.string.events_count, count),
            style = TextStyle(
                color = Color.Black,
                fontSize = 18.sp,
                fontWeight = FontWeight.Bold,
            ),
        )

        Icon(
            modifier = Modifier
                .align(Alignment.CenterEnd)
                .size(32.dp)
                .customClickable(onClick = onDeleteClick),
            imageVector = Icons.Filled.Delete,
            tint = Color.Red,
            contentDescription = "",
        )
    }
}

@Preview
@Composable
private fun ParamCardPreview() {
    ParamCard(
        modifier = Modifier.fillMaxWidth(),
        param = ParamUiData(id = 0, title = "text: Sign in", locale = "eng"),
        onDeleteClick = {},
    )
}

@Preview
@Composable
private fun EventsCardPreview() {
    EventsCard(
        modifier = Modifier.fillMaxWidth(),
        count = 24,
        onDeleteClick = {},
    )
}
