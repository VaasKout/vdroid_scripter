package com.vision.scripter.prefs

import android.content.Context
import androidx.datastore.core.DataStore
import androidx.datastore.preferences.core.Preferences
import androidx.datastore.preferences.core.edit
import androidx.datastore.preferences.core.stringPreferencesKey
import androidx.datastore.preferences.preferencesDataStore
import com.vision.scripter.coroutines.api.DispatchersFactory
import com.vision.scripter.prefs.api.DataStoreRepository
import dagger.hilt.android.qualifiers.ApplicationContext
import kotlinx.coroutines.flow.firstOrNull
import kotlinx.coroutines.flow.map
import kotlinx.coroutines.withContext
import javax.inject.Inject
import javax.inject.Singleton

private const val SERIAL_NUMBER_KEY = "serial_number"
private const val URL_KEY = "url"

@Singleton
class DataStoreRepositoryImpl @Inject constructor(
    @ApplicationContext context: Context,
    private val dispatchersFactory: DispatchersFactory,
) : DataStoreRepository {

    private val Context.dataStore: DataStore<Preferences> by preferencesDataStore(
        name = "user_prefs",
    )
    private val dataStore = context.dataStore

    override suspend fun saveSerialNumber(serialNumber: String) {
        withContext(dispatchersFactory.io) {
            dataStore.edit { it[stringPreferencesKey(SERIAL_NUMBER_KEY)] = serialNumber }
        }
    }

    override suspend fun getSerialNumber(): String {
        return withContext(dispatchersFactory.io) {
            dataStore.data.map { it[stringPreferencesKey(SERIAL_NUMBER_KEY)] }
                .firstOrNull()
                .orEmpty()
        }
    }

    override suspend fun setServerUrl(url: String) {
        withContext(dispatchersFactory.io) {
            dataStore.edit { it[stringPreferencesKey(URL_KEY)] = url }
        }
    }

    override suspend fun getServerUrl(): String {
        return withContext(dispatchersFactory.io) {
            dataStore.data.map { it[stringPreferencesKey(URL_KEY)] }
                .firstOrNull()
                .orEmpty()
        }
    }
}