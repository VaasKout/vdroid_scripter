package com.vision.scripter.libraryitems.data

import com.vision.scripter.coroutines.api.CoroutineScopeFactory
import com.vision.scripter.data.api.ScripterDataSource
import com.vision.scripter.data.api.models.AdbDevice
import com.vision.scripter.data.api.models.SessionStatus
import com.vision.scripter.data.api.models.Step
import com.vision.scripter.library.state.LibraryType
import com.vision.scripter.network.api.ApiResponse
import com.vision.scripter.network.api.NetworkError
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Job
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.isActive
import kotlinx.coroutines.launch
import java.util.concurrent.ConcurrentHashMap
import javax.inject.Inject
import javax.inject.Singleton
import kotlin.coroutines.coroutineContext
import kotlin.time.Duration.Companion.milliseconds

private const val NoConnectionText = "no connection"
private const val StartedText = "running"

@Singleton
class RunnerRepository @Inject constructor(
    coroutineScopeFactory: CoroutineScopeFactory,
    private val scripterDataSource: ScripterDataSource,
) {

    private val coroutineScope: CoroutineScope =
        coroutineScopeFactory.createIoScope("runner_repository")

    private val _runsFlow = MutableStateFlow<Map<String, Run>>(mapOf())
    fun observeRuns(): StateFlow<Map<String, Run>> = _runsFlow.asStateFlow()

    private val pollingJobs = ConcurrentHashMap<String, Job>()

    suspend fun run(
        type: LibraryType,
        name: String,
        device: AdbDevice,
    ): ApiResponse<Unit> {
        val result = when (type) {
            LibraryType.ROUTES -> scripterDataSource.runRoute(serial = device.serial, name = name)
            LibraryType.ACTIONS -> scripterDataSource.queueSteps(
                serial = device.serial,
                steps = listOf(Step(event = name)),
            )

            LibraryType.IMAGES -> ApiResponse.Error(NetworkError.ServerError("images can't run"))
        }
        if (result is ApiResponse.Error) return result

        val run = Run(
            type = type,
            name = name,
            serial = device.serial,
            deviceLabel = device.label(),
            status = SessionStatus.Running(StartedText),
        )
        _runsFlow.update { it + (device.serial to run) }
        startPolling(device.serial)
        return result
    }

    fun dismiss(serial: String) {
        pollingJobs.remove(serial)?.cancel()
        _runsFlow.update { it - serial }
    }

    private fun startPolling(serial: String) {
        pollingJobs.remove(serial)?.cancel()
        pollingJobs[serial] = coroutineScope.launch {
            pollUntilDone(serial)
            coroutineContext[Job]?.let { pollingJobs.remove(serial, it) }
        }
    }

    private suspend fun pollUntilDone(serial: String) {
        while (coroutineContext.isActive) {
            delay(500.milliseconds)
            val status = pollStatus(serial)
            updateStatus(serial, status)
            if (status is SessionStatus.Running) continue
            if (status is SessionStatus.Error) return
            _runsFlow.update { it - serial }
            return
        }
    }

    private suspend fun pollStatus(serial: String): SessionStatus {
        return when (val result = scripterDataSource.getSessionStatus(serial)) {
            is ApiResponse.Success -> result.data
            is ApiResponse.Error -> SessionStatus.Error(
                result.error.text().ifEmpty { NoConnectionText },
            )
        }
    }

    private fun updateStatus(serial: String, status: SessionStatus) {
        _runsFlow.update { runs ->
            val run = runs[serial] ?: return@update runs
            runs + (serial to run.copy(status = status))
        }
    }
}
