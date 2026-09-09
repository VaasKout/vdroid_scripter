package com.vision.scripter.network.impl

import com.vision.scripter.coroutines.api.DispatchersFactory
import com.vision.scripter.network.api.ApiResponse
import com.vision.scripter.network.api.NetworkClient
import com.vision.scripter.network.api.NetworkError
import com.vision.scripter.prefs.api.DataStoreRepository
import io.ktor.client.HttpClient
import io.ktor.client.engine.cio.CIO
import io.ktor.client.plugins.HttpTimeout
import io.ktor.client.plugins.contentnegotiation.ContentNegotiation
import io.ktor.client.request.delete
import io.ktor.client.request.get
import io.ktor.client.request.post
import io.ktor.client.request.setBody
import io.ktor.client.statement.bodyAsText
import io.ktor.http.ContentType
import io.ktor.http.contentType
import kotlinx.coroutines.withContext
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class NetworkClientImpl @Inject constructor(
    private val dispatchersFactory: DispatchersFactory,
    private val dataStoreRepository: DataStoreRepository,
) : NetworkClient {

    var client = HttpClient(CIO) {
        install(HttpTimeout)
        install(ContentNegotiation)
    }

    override suspend fun get(path: String): ApiResponse<String> {
        return withContext(dispatchersFactory.io) {
            val url = buildUrl(path)
            if (url.isEmpty()) {
                return@withContext ApiResponse.Error(error = NetworkError.NoUrlError)
            }
            try {
                val response = client.get(url) {
                    contentType(ContentType.Application.Json)
                }
                if (response.status.value !in (200..299)) {
                    throw Exception(response.bodyAsText())
                }
                ApiResponse.Success(response.bodyAsText())
            } catch (e: Exception) {
                e.printStackTrace()
                val error = NetworkError.ServerError(msg = "Error: ${e.message}")
                ApiResponse.Error(error = error)
            }
        }
    }

    override suspend fun post(path: String, body: String): ApiResponse<String> {
        return withContext(dispatchersFactory.io) {
            val url = buildUrl(path)
            if (url.isEmpty()) {
                return@withContext ApiResponse.Error(error = NetworkError.NoUrlError)
            }
            try {
                val response = client.post(url) {
                    setBody(body)
                    contentType(ContentType.Application.Json)
                }
                if (response.status.value !in (200..299)) {
                    throw Exception(response.bodyAsText())
                }
                ApiResponse.Success(response.bodyAsText())
            } catch (e: Exception) {
                e.printStackTrace()
                val error = NetworkError.ServerError(msg = "Error: ${e.message}")
                ApiResponse.Error(error = error)
            }
        }
    }

    override suspend fun delete(path: String): ApiResponse<String> {
        return withContext(dispatchersFactory.io) {
            val url = buildUrl(path)
            if (url.isEmpty()) {
                return@withContext ApiResponse.Error(error = NetworkError.NoUrlError)
            }
            try {
                val response = client.delete(url) {
                    contentType(ContentType.Application.Json)
                }
                if (response.status.value !in (200..299)) {
                    throw Exception(response.bodyAsText())
                }
                ApiResponse.Success(response.bodyAsText())
            } catch (e: Exception) {
                e.printStackTrace()
                val error = NetworkError.ServerError(msg = "Error: ${e.message}")
                ApiResponse.Error(error = error)
            }
        }
    }

    private suspend fun buildUrl(path: String): String {
        val baseUrl = dataStoreRepository.getServerUrl()
        if (baseUrl.isEmpty()) return ""
        return buildString {
            append(baseUrl)
            if (baseUrl.last() != '/') append('/')
            if (path.isNotEmpty() && path.first() == '/') {
                append(path.substring(1))
            } else {
                append(path)
            }
        }
    }
}
