package com.vision.scripter.libraryitems.state

import com.vision.scripter.coroutines.api.CoroutineScopeFactory
import com.vision.scripter.data.api.ScripterDataSource
import com.vision.scripter.data.api.models.AdbDevice
import com.vision.scripter.data.api.models.SessionStatus
import com.vision.scripter.library.state.LibraryType
import com.vision.scripter.libraryitems.data.RunnerRepository
import com.vision.scripter.libraryitems.data.text
import com.vision.scripter.libraryitems.ui.LibraryItemsUiCommand
import com.vision.scripter.libraryitems.ui.LibraryItemsUiState
import com.vision.scripter.libraryitems.ui.LibraryItemsUiStateHolder
import com.vision.scripter.network.api.ApiResponse
import com.vision.scripter.prefs.api.DataStoreRepository
import com.vision.scripter.ui.CommandFlow
import com.vision.scripter.ui.states.LoadingState
import dagger.hilt.android.scopes.ViewModelScoped
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.cancel
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharingStarted
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.launchIn
import kotlinx.coroutines.flow.map
import kotlinx.coroutines.flow.onEach
import kotlinx.coroutines.flow.stateIn
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject
import kotlin.time.Duration.Companion.milliseconds

@ViewModelScoped
internal class LibraryItemsInteractor @Inject constructor(
    coroutineScopeFactory: CoroutineScopeFactory,
    private val scripterDataSource: ScripterDataSource,
    private val dataStoreRepository: DataStoreRepository,
    private val runnerRepository: RunnerRepository,
    private val uiStateMapper: LibraryItemsUiStateMapper,
) : LibraryItemsUiStateHolder {

    private val _stateFlow = MutableStateFlow(LibraryItemsState())
    private val stateFlow: StateFlow<LibraryItemsState> = _stateFlow.asStateFlow()

    private val coroutineScope: CoroutineScope =
        coroutineScopeFactory.createBackgroundScope("library_items_interactor")

    private val currentState: LibraryItemsState
        get() = _stateFlow.value

    override val uiStateFlow: StateFlow<LibraryItemsUiState> = stateFlow
        .map(uiStateMapper::map)
        .stateIn(coroutineScope, SharingStarted.Eagerly, LibraryItemsUiState())

    override val uiCommandsFlow: CommandFlow<LibraryItemsUiCommand> = CommandFlow(coroutineScope)

    init {
        startReactiveStreams()
    }

    private fun startReactiveStreams() {
        runnerRepository.observeRuns().onEach { runs ->
            _stateFlow.update { it.copy(runs = runs) }
        }.launchIn(coroutineScope)
    }

    override fun init(type: LibraryType) {
        _stateFlow.update { it.copy(type = type) }
        onLoadData(onStart = true)
    }

    override fun onLoadData(onStart: Boolean) {
        coroutineScope.launch {
            _stateFlow.update {
                it.copy(
                    loadingState = if (onStart) LoadingState.LoadingOnStart
                    else LoadingState.RefreshLoading,
                )
            }

            when (val names = loadNames(currentState.type)) {
                null -> uiCommandsFlow.tryEmit(LibraryItemsUiCommand.ShowNetworkError)
                else -> _stateFlow.update { it.copy(names = names) }
            }

            if (!onStart) delay(500.milliseconds)
            _stateFlow.update {
                it.copy(loadingState = LoadingState.None)
            }
        }
    }

    override fun onDeleteItem(name: String) {
        _stateFlow.update { it.copy(itemToDelete = name) }
    }

    override fun onDismissDelete() {
        _stateFlow.update { it.copy(itemToDelete = null) }
    }

    override fun onConfirmDelete() {
        val name = currentState.itemToDelete ?: return
        coroutineScope.launch {
            val deleted = when (currentState.type) {
                LibraryType.IMAGES -> scripterDataSource.deleteImage(name)
                LibraryType.ACTIONS -> scripterDataSource.deleteAction(name)
                LibraryType.ROUTES -> scripterDataSource.deleteRoute(name)
            }
            if (!deleted) uiCommandsFlow.tryEmit(LibraryItemsUiCommand.ShowNetworkError)
            onDismissDelete()
            onLoadData(onStart = false)
        }
    }

    override fun onPlayClicked(name: String) {
        if (currentState.type == LibraryType.IMAGES) return
        _stateFlow.update { it.copy(picker = DevicePicker(itemName = name)) }
        coroutineScope.launch {
            val lastSerial = dataStoreRepository.getSerialNumber()
            val devices = loadDevices()
            if (devices == null) {
                _stateFlow.update { it.copy(picker = null) }
                uiCommandsFlow.tryEmit(LibraryItemsUiCommand.ShowNetworkError)
                return@launch
            }
            val pickerDevices = devices.map { device ->
                PickerDevice(device = device, status = sessionStatus(device.serial))
            }
            _stateFlow.update { state ->
                val picker = state.picker ?: return@update state
                state.copy(
                    picker = picker.copy(
                        isLoading = false,
                        devices = pickerDevices,
                        lastSerial = lastSerial,
                    )
                )
            }
        }
    }

    override fun onDeviceChosen(serial: String) {
        val picker = currentState.picker ?: return
        val chosen = picker.devices.firstOrNull { it.device.serial == serial } ?: return
        if (chosen.status is SessionStatus.Running) return
        _stateFlow.update { it.copy(picker = null) }
        coroutineScope.launch {
            dataStoreRepository.saveSerialNumber(serial)
            val result = runnerRepository.run(
                type = currentState.type,
                name = picker.itemName,
                device = chosen.device,
            )
            if (result is ApiResponse.Error) {
                uiCommandsFlow.tryEmit(LibraryItemsUiCommand.ShowError(result.error.text()))
            }
        }
    }

    override fun onPickerDismissed() {
        _stateFlow.update { it.copy(picker = null) }
    }

    override fun onRunDismissed(name: String) {
        val run = currentState.runs.values.firstOrNull {
            it.type == currentState.type && it.name == name
        } ?: return
        if (run.status !is SessionStatus.Error) return
        runnerRepository.dismiss(run.serial)
    }

    private suspend fun loadNames(type: LibraryType): List<String>? {
        if (type == LibraryType.ROUTES) {
            return when (val result = scripterDataSource.getRoutes()) {
                is ApiResponse.Success -> result.data
                is ApiResponse.Error -> null
            }
        }
        return when (val result = scripterDataSource.getLibrary()) {
            is ApiResponse.Success -> {
                if (type == LibraryType.IMAGES) result.data.images else result.data.actions
            }

            is ApiResponse.Error -> null
        }
    }

    private suspend fun loadDevices(): List<AdbDevice>? {
        return when (val result = scripterDataSource.getDevices()) {
            is ApiResponse.Success -> result.data
            is ApiResponse.Error -> null
        }
    }

    private suspend fun sessionStatus(serial: String): SessionStatus {
        return when (val result = scripterDataSource.getSessionStatus(serial)) {
            is ApiResponse.Success -> result.data
            is ApiResponse.Error -> SessionStatus.Closed
        }
    }

    fun clear() {
        coroutineScope.cancel()
    }
}
