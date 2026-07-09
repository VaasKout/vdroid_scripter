package com.vision.scripter.ui

import androidx.compose.foundation.Indication
import androidx.compose.foundation.clickable
import androidx.compose.foundation.interaction.MutableInteractionSource
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.ripple
import androidx.compose.runtime.Composable
import androidx.compose.runtime.remember
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Shape
import androidx.compose.ui.unit.dp

@Composable
fun Modifier.conditional(
    flag: Boolean,
    apply: Modifier.() -> Modifier,
): Modifier {
    return if (flag) this.apply() else this
}


@Composable
fun Modifier.customClickable(
    shape: Shape = RoundedCornerShape(16.dp),
    indication: Indication? = ripple(bounded = true),
    onClick: () -> Unit,
): Modifier {
    return this
        .clip(shape)
        .clickable(
            indication = indication,
            interactionSource = remember { MutableInteractionSource() },
            onClick = onClick,
        )
}
