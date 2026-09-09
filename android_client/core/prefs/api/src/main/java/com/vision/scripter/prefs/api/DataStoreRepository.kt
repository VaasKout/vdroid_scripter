package com.vision.scripter.prefs.api

interface DataStoreRepository {
    suspend fun saveSerialNumber(serialNumber: String)
    suspend fun getSerialNumber(): String
    suspend fun setServerUrl(url: String)
    suspend fun getServerUrl(): String
}
