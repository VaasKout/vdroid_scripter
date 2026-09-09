package com.vision.scripter.network.api

sealed class ApiResponse<out T> {
    data class Success<out T>(val data: T) : ApiResponse<T>()
    data class Error(val error: NetworkError) : ApiResponse<Nothing>()
}

sealed class NetworkError {
    object NoUrlError : NetworkError()
    data class ServerError(val msg: String) : NetworkError()
}
