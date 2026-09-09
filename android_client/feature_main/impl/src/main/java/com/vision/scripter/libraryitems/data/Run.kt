package com.vision.scripter.libraryitems.data

import com.vision.scripter.data.api.models.AdbDevice
import com.vision.scripter.data.api.models.SessionStatus
import com.vision.scripter.library.state.LibraryType
import com.vision.scripter.network.api.NetworkError

data class Run(
    val type: LibraryType,
    val name: String,
    val serial: String,
    val deviceLabel: String,
    val status: SessionStatus,
)

fun AdbDevice.label(): String {
    val name = marketingName.ifEmpty { model }.ifEmpty { device }
    if (name.isEmpty()) return serial
    return "$name ($serial)"
}

fun NetworkError.text(): String = when (this) {
    is NetworkError.ServerError -> msg
    is NetworkError.NoUrlError -> ""
}
