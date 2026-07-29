package com.vision.scripter.streaming.impl.domain

import com.vision.scripter.data.api.models.Event
import com.vision.scripter.data.api.models.Parameter
import com.vision.scripter.streaming.impl.screen.state.DEFAULT_TIMEOUT
import com.vision.scripter.streaming.impl.screen.state.NEW_SCRIPTS_NODE

data class Record(
    val controlRecording: Boolean = false,
    val recordName: String = "",
    val node: String = NEW_SCRIPTS_NODE,
    val params: List<Parameter> = listOf(),
    val events: List<Event> = listOf(),
    val timeout: Int = DEFAULT_TIMEOUT,
)