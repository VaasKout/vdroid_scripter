package com.vision.scripter.network.api

interface NetworkClient {
    suspend fun get(path: String): ApiResponse<String>
    suspend fun post(path: String, body: String): ApiResponse<String>
    suspend fun delete(path: String): ApiResponse<String>
}
