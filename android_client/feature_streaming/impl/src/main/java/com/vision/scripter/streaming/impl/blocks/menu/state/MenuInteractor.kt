package com.vision.scripter.streaming.impl.blocks.menu.state

import com.vision.scripter.coroutines.api.CoroutineScopeFactory
import com.vision.scripter.data.api.ScripterDataSource
import com.vision.scripter.data.api.models.adjustToServer
import com.vision.scripter.network.api.ApiResponse
import com.vision.scripter.streaming.impl.blocks.menu.ui.MenuUiCommand
import com.vision.scripter.streaming.impl.blocks.menu.ui.MenuUiState
import com.vision.scripter.streaming.impl.blocks.menu.ui.MenuUiStateHolder
import com.vision.scripter.streaming.impl.data.CvRepository
import com.vision.scripter.streaming.impl.data.ItemType
import com.vision.scripter.streaming.impl.data.RecordRepository
import com.vision.scripter.streaming.impl.data.VideoStreamerRepository
import com.vision.scripter.streaming.impl.screen.StreamingEvent
import com.vision.scripter.streaming.impl.screen.StreamingEventsHolder
import com.vision.scripter.ui.CommandFlow
import dagger.hilt.android.scopes.ViewModelScoped
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.cancel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharingStarted
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.combine
import kotlinx.coroutines.flow.launchIn
import kotlinx.coroutines.flow.onEach
import kotlinx.coroutines.flow.stateIn
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject

@ViewModelScoped
class MenuInteractor @Inject constructor(
    coroutineScopeFactory: CoroutineScopeFactory,
    uiStateMapper: MenuUiStateMapper,
    private val scripterDataSource: ScripterDataSource,
    private val cvRepository: CvRepository,
    private val recordRepository: RecordRepository,
    private val videoRepository: VideoStreamerRepository,
    private val eventRepository: StreamingEventsHolder,
) : MenuUiStateHolder {

    private val coroutineScope: CoroutineScope =
        coroutineScopeFactory.createBackgroundScope("menu_interactor")

    override val uiCommandsFlow: CommandFlow<MenuUiCommand> = CommandFlow(coroutineScope)

    private val _menuState = MutableStateFlow(MenuState())
    private val menuState = _menuState.asStateFlow()

    private val _dialogState = MutableStateFlow<DialogState>(DialogState.None)
    private val dialogState = _dialogState.asStateFlow()

    private val serial: String
        get() = _menuState.value.serial

    override val uiStateFlow: StateFlow<MenuUiState> = combine(
        menuState,
        dialogState,
    ) { menuState, dialogState ->
        uiStateMapper.map(
            state = menuState,
            dialogState = dialogState,
        )
    }.stateIn(coroutineScope, SharingStarted.Eagerly, MenuUiState())

    init {
        startReactiveStreams()
    }

    override fun init(serial: String) {
        _menuState.update {
            it.copy(serial = serial)
        }
    }

    private fun startReactiveStreams() {
        recordRepository.observeRecord().onEach { record ->
            _menuState.update {
                val type = it.type
                if (type !is MenuType.CustomAction) return@update it
                it.copy(type = type.copy(recording = record.recording))
            }
        }.launchIn(coroutineScope)
    }

    override fun onAddClicked() {
        _dialogState.update { DialogState.AddItem }
    }

    override fun onAddItemConfirmed(name: String, itemType: ItemType) {
        hideDialog()
        val trimmedName = name.trim()
        if (trimmedName.isEmpty() || itemType == ItemType.NONE) return

        recordRepository.initData(name = trimmedName, itemType = itemType)
        if (itemType == ItemType.IMAGE) {
            _menuState.update { it.copy(type = MenuType.SelectingCV()) }
            refreshRectangles()
            return
        }

        if (itemType == ItemType.ACTION) {
            _menuState.update { it.copy(type = MenuType.CustomAction()) }
        }
    }

    override fun onRectanglesClicked() {
        val type = _menuState.value.type
        if (type !is MenuType.Usual || type.rectsAreLoading) return
        if (type.rectsShown) {
            _menuState.update { it.copy(type = type.copy(rectsShown = false)) }
            cvRepository.clearOverlay()
            return
        }
        _menuState.update { it.copy(type = type.copy(rectsShown = true, scanShown = false)) }
        refreshRectangles()
    }

    override fun onRefreshRectanglesClicked() {
        if (rectanglesLoading()) return
        refreshRectangles()
    }

    override fun onScanClicked() {
        val type = _menuState.value.type
        if (type !is MenuType.Usual || type.scanning) return
        if (type.scanShown) {
            _menuState.update { it.copy(type = type.copy(scanShown = false)) }
            cvRepository.clearOverlay()
            return
        }
        _dialogState.update { DialogState.Scan }
    }

    override fun onScanConfirmed(locale: String, includeImages: Boolean) {
        hideDialog()
        val type = _menuState.value.type
        if (type !is MenuType.Usual) return
        _menuState.update { it.copy(type = type.copy(scanning = true, rectsShown = false)) }
        coroutineScope.launch {
            val found = scan(locale = locale, includeImages = includeImages)
            _menuState.update { state ->
                val current = state.type
                if (current !is MenuType.Usual) return@update state
                state.copy(type = current.copy(scanning = false, scanShown = found))
            }
            if (!found) eventRepository.sendEvent(StreamingEvent.ShowNetworkError)
        }
    }

    override fun onExpandClicked() {
        val type = _menuState.value.type
        if (type is MenuType.Usual) {
            _menuState.update { it.copy(type = type.copy(expanded = !type.expanded)) }
        }
    }

    override fun onRecordingClicked() {
        val state = _menuState.value
        if (state.type !is MenuType.CustomAction) return
        recordRepository.switchRecording()
    }

    override fun onSaveClicked() {
        when (_menuState.value.type) {
            is MenuType.SelectingCV -> saveImage()
            is MenuType.CustomAction -> saveAction()
            else -> Unit
        }
    }

    override fun onCancelClicked() {
        _menuState.update { it.copy(type = MenuType.Usual(expanded = true)) }
        coroutineScope.launch {
            recordRepository.clear()
            dropState()
        }
    }

    override fun onExitClicked() {
        uiCommandsFlow.tryEmit(MenuUiCommand.ExitCommand)
    }

    override fun onDialogDismissed() {
        hideDialog()
    }

    private fun saveImage() {
        _menuState.update { it.copy(type = MenuType.Usual(expanded = true)) }
        coroutineScope.launch {
            val screenSizes = videoRepository.observeScreenSizes().value
            val selected = cvRepository.observeSelectedRectangles().value.firstOrNull()
            if (screenSizes == null || selected == null) {
                recordRepository.clear()
                dropState()
                return@launch
            }

            val saved = recordRepository.saveImage(
                serial = serial,
                rectangle = selected.adjustToServer(screenSizes),
            )
            if (saved) {
                eventRepository.sendEvent(StreamingEvent.ShowItemSavedSnackbar)
            } else {
                eventRepository.sendEvent(StreamingEvent.ShowNetworkError)
            }
            recordRepository.clear()
            dropState()
        }
    }

    private fun saveAction() {
        _menuState.update { it.copy(type = MenuType.Usual(expanded = true)) }
        coroutineScope.launch {
            val screenSizes = videoRepository.observeScreenSizes().value
            val events = recordRepository.observeRecord().value.events
            if (screenSizes == null || events.isEmpty()) {
                recordRepository.clear()
                dropState()
                return@launch
            }

            val saved = recordRepository.saveAction(screenSizes)
            if (saved) {
                eventRepository.sendEvent(StreamingEvent.ShowItemSavedSnackbar)
            } else {
                eventRepository.sendEvent(StreamingEvent.ShowNetworkError)
            }
            recordRepository.clear()
            dropState()
        }
    }

    private fun refreshRectangles() {
        coroutineScope.launch {
            val screenSizes = videoRepository.observeScreenSizes().value ?: return@launch
            setRectanglesLoading(true)
            val loaded = cvRepository.refreshRectangles(serial = serial, screenSizes = screenSizes)
            setRectanglesLoading(false)
            if (loaded) return@launch
            hideRectanglesOverlay()
            eventRepository.sendEvent(StreamingEvent.ShowNetworkError)
        }
    }

    private fun rectanglesLoading(): Boolean = when (val type = _menuState.value.type) {
        is MenuType.Usual -> type.rectsAreLoading
        is MenuType.SelectingCV -> type.rectsLoading
        else -> false
    }

    private fun setRectanglesLoading(isLoading: Boolean) {
        _menuState.update { state ->
            when (val type = state.type) {
                is MenuType.Usual -> state.copy(type = type.copy(rectsAreLoading = isLoading))
                is MenuType.SelectingCV -> state.copy(type = type.copy(rectsLoading = isLoading))
                else -> state
            }
        }
    }

    private fun hideRectanglesOverlay() {
        _menuState.update { state ->
            val type = state.type
            if (type !is MenuType.Usual) return@update state
            state.copy(type = type.copy(rectsShown = false))
        }
    }

    private suspend fun scan(locale: String, includeImages: Boolean): Boolean {
        val screenSizes = videoRepository.observeScreenSizes().value ?: return false
        val images = if (includeImages) libraryImages() ?: return false else listOf()
        return cvRepository.scan(
            serial = serial,
            images = images,
            locale = locale,
            screenSizes = screenSizes,
        )
    }

    private suspend fun libraryImages(): List<String>? {
        return when (val result = scripterDataSource.getLibrary()) {
            is ApiResponse.Success -> result.data.images
            is ApiResponse.Error -> null
        }
    }

    private fun dropState() {
        cvRepository.clearSelectedRectangles()
        cvRepository.clearOverlay()
    }

    private fun hideDialog() {
        _dialogState.update { DialogState.None }
    }

    fun clear() {
        _menuState.update { MenuState() }
        _dialogState.update { DialogState.None }
        coroutineScope.cancel()
    }
}
