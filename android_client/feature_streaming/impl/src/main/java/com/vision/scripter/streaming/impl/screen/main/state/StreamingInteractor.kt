package com.vision.scripter.streaming.impl.screen.main.state

import android.view.MotionEvent
import android.view.MotionEvent.ACTION_DOWN
import android.view.MotionEvent.ACTION_UP
import android.view.Surface
import androidx.core.net.toUri
import com.vision.scripter.coroutines.api.CoroutineScopeFactory
import com.vision.scripter.data.api.ControlStreamer
import com.vision.scripter.data.api.ScripterRepository
import com.vision.scripter.data.api.models.Event
import com.vision.scripter.data.api.models.Parameter
import com.vision.scripter.data.api.models.RectangleWithText
import com.vision.scripter.data.api.models.Script
import com.vision.scripter.data.api.models.StreamingData
import com.vision.scripter.data.api.models.adjustToClient
import com.vision.scripter.data.api.models.contains
import com.vision.scripter.network.api.ApiResponse
import com.vision.scripter.prefs.api.DataStoreRepository
import com.vision.scripter.streaming.impl.blocks.menu.state.MenuEvent
import com.vision.scripter.streaming.impl.blocks.menu.state.MenuInteractor
import com.vision.scripter.streaming.impl.screen.main.ui.StreamingUiCommand
import com.vision.scripter.streaming.impl.screen.main.ui.StreamingUiState
import com.vision.scripter.streaming.impl.screen.main.ui.StreamingUiStateHolder
import com.vision.scripter.streaming.impl.usecases.CvUseCase
import com.vision.scripter.streaming.impl.usecases.VideoUseCase
import com.vision.scripter.ui.CommandFlow
import com.vision.scripter.ui.states.LoadingState
import dagger.hilt.android.scopes.ViewModelScoped
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Job
import kotlinx.coroutines.cancel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharedFlow
import kotlinx.coroutines.flow.SharingStarted
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.combine
import kotlinx.coroutines.flow.launchIn
import kotlinx.coroutines.flow.onEach
import kotlinx.coroutines.flow.stateIn
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import kotlinx.coroutines.sync.Mutex
import kotlinx.coroutines.sync.withLock
import javax.inject.Inject

@ViewModelScoped
class StreamingInteractor @Inject constructor(
    coroutineScopeFactory: CoroutineScopeFactory,
    private val scripterRepository: ScripterRepository,
    private val dataStoreRepository: DataStoreRepository,
    private val uiStateMapper: StreamingUiStateMapper,
    private val videoUseCase: VideoUseCase,
    private val controlStreamer: ControlStreamer,
    private val cvUseCase: CvUseCase,
    val menuInteractor: MenuInteractor,
) : StreamingUiStateHolder {

    private val coroutineScope: CoroutineScope =
        coroutineScopeFactory.createBackgroundScope("streaming_interactor")

    private val _stateFlow = MutableStateFlow(StreamingState())
    private val stateFlow: SharedFlow<StreamingState> = _stateFlow.asStateFlow()

    private val currentState: StreamingState
        get() = _stateFlow.value

    override val uiStateFlow: StateFlow<StreamingUiState?> = combine(
        stateFlow,
        menuInteractor.observeMenuState(),
        menuInteractor.observeDialogState(),
    ) { streamingState, menuState, dialogState ->
        uiStateMapper.map(
            state = streamingState,
            menuState = menuState,
            dialogState = dialogState,
        )
    }.stateIn(coroutineScope, SharingStarted.Eagerly, null)

    override val uiCommandsFlow: CommandFlow<StreamingUiCommand> = CommandFlow(coroutineScope)

    private val mutex = Mutex()

    @Volatile
    private var streamJob: Job? = null

    private var startRecordingTime = 0L

    override fun initArgs(serial: String) {
        _stateFlow.update {
            it.copy(serial = serial)
        }
    }

    init {
        cvUseCase.observeRectangles(coroutineScope).onEach { rectangles ->
            _stateFlow.update {
                it.copy(cvRectangles = rectangles)
            }
        }.launchIn(coroutineScope)

        cvUseCase.observeSelectedRectangles().onEach { selected ->
            _stateFlow.update {
                it.copy(selectedRectangles = selected)
            }
        }.launchIn(coroutineScope)

        menuInteractor.observeEvents().onEach(::handleMenuEvent).launchIn(coroutineScope)
    }

    private suspend fun handleMenuEvent(event: MenuEvent) {
        when (event) {
            is MenuEvent.NextCvMode -> {
                if (event.cvMode == CVMode.NO_CV) {
                    cvUseCase.restoreSelectedRectangles()
                } else {
                    cvUseCase.snapshotSelectedRectangles()
                }
                cvUseCase.nextCvMode(event.cvMode)
            }

            is MenuEvent.SaveClicked -> {
                if (event.param == null) {
                    saveScript()
                    return
                }
                if (event.param.type == TEMPLATE) {
                    val screenSizes = videoUseCase.observeScreenSizes().value ?: return
                    cvUseCase.nextCvMode(CVMode.NO_CV)
                    val templateName = "${currentState.record.params.size + 1}"
                    cvUseCase.saveSelectedRectangle(
                        serial = currentState.serial,
                        node = currentState.record.node,
                        name = currentState.record.recordName,
                        value = templateName,
                        screenSizes = screenSizes,
                    )
                    addParam(Parameter(type = TEMPLATE, value = templateName))
                    return
                }
                if (event.param.type == YOLO_CLASS) {
                    val label =
                        cvUseCase.observeSelectedRectangles().value.firstOrNull()?.label.orEmpty()
                    addParam(Parameter(type = YOLO_CLASS, value = label))
                    return
                }
                if (event.param.type.isNotEmpty()) {
                    addParam(event.param)
                }
            }

            is MenuEvent.RecordCancelled -> {
                _stateFlow.update { it.copy(record = StreamingState.Record()) }
                cvUseCase.clearAllRectangles()
            }

            is MenuEvent.FindText -> {
                val screenSizes = videoUseCase.observeScreenSizes().value ?: return
                val found = cvUseCase.findTextRectangles(
                    serial = currentState.serial,
                    text = event.text,
                    locale = event.locale,
                    screenSizes = screenSizes,
                )
                if (!found) {
                    uiCommandsFlow.tryEmit(StreamingUiCommand.ShowNetworkError)
                    return
                }
                menuInteractor.onTextSearchSuccess(text = event.text, locale = event.locale)
            }

            is MenuEvent.KeyboardInit -> {
                val result = scripterRepository.resetKeyboard(
                    serial = currentState.serial,
                    locale = currentState.record.locale,
                )
                if (result is ApiResponse.Success) {
                    setupKeyboardRects(result.data)
                }
            }

            is MenuEvent.EditKeyboardButton -> {
                val screenSizes = videoUseCase.observeScreenSizes().value ?: return
                val success = cvUseCase.editKeyboardSelectedRectangle(
                    serial = currentState.serial,
                    locale = currentState.record.locale,
                    oldName = event.oldKey,
                    newName = event.newKey,
                    screenSizes = screenSizes,
                )
                cvUseCase.clearSelectedRectangles()
                if (success) {
                    getOrResetKeyboard()
                    return
                }
                uiCommandsFlow.tryEmit(StreamingUiCommand.ShowNetworkError)
            }

            is MenuEvent.SaveLocale -> {
                _stateFlow.update {
                    it.copy(record = it.record.copy(locale = event.locale))
                }
                getOrResetKeyboard()
            }

            is MenuEvent.SaveRecordName -> _stateFlow.update {
                it.copy(record = it.record.copy(recordName = event.name))
            }

            is MenuEvent.TimeoutSaved -> _stateFlow.update {
                it.copy(record = it.record.copy(timeout = event.timeout))
            }
        }
    }

    override fun onLoadData(onStart: Boolean) {
        coroutineScope.launch {
            _stateFlow.update {
                it.copy(loadingState = LoadingState.LoadingOnStart)
            }
            val result = getStreamingData()
            if (result is ApiResponse.Error) {
                uiCommandsFlow.tryEmit(StreamingUiCommand.ShowNetworkError)
                _stateFlow.update {
                    it.copy(loadingState = LoadingState.None)
                }
            }
        }
    }

    override fun onVideoSurfaceCreated(
        surfaceWidth: Int,
        surfaceHeight: Int,
        newSurface: Surface,
    ) {
        streamJob = coroutineScope.launch {
            val streamingData = currentState.streamingData ?: return@launch
            val videoConnected = videoUseCase.initConnection(
                host = currentState.streamingHost,
                port = streamingData.videoPort.toInt(),
                newSurface = newSurface,
                surfaceWidth = surfaceWidth,
                surfaceHeight = surfaceHeight,
            )

            val controlConnected = controlStreamer.initConnection(
                host = currentState.streamingHost,
                port = streamingData.controlPort.toInt(),
            )

            val cvConnected = cvUseCase.initConnection(
                host = currentState.streamingHost,
                port = streamingData.cvPort.toInt(),
            )

            val screenSizes = videoUseCase.observeScreenSizes().value
            val connectionEstablished =
                videoConnected && controlConnected && cvConnected && screenSizes != null

            _stateFlow.update {
                it.copy(loadingState = LoadingState.None)
            }

            if (!connectionEstablished) {
                _stateFlow.update {
                    it.copy(streamingData = null)
                }
                return@launch
            }

            launch {
                videoUseCase.decodeFramesInLoop(
                    mimeType = currentState.videoCodec.mimeType,
                )
            }
            cvUseCase.decodeRectanglesInLoop(
                screenSizes = screenSizes,
            )
        }
    }

    override fun onVideoSurfaceDestroyed() {
        closeStreams()
    }

    fun closeStreams() {
        if (streamJob?.isActive == true) {
            streamJob?.cancel()
            streamJob = null
        }
        videoUseCase.stop()
        controlStreamer.close()
        cvUseCase.close()
    }

    fun clear() {
        coroutineScope.cancel()
    }

    override fun onTouchEvent(
        viewWidth: Int,
        viewHeight: Int,
        event: MotionEvent?,
    ) {
        if (event == null) return
        coroutineScope.launch {
            mutex.withLock {
                try {
                    val menuState = menuInteractor.observeMenuState().value
                    if (menuState is MenuState.SelectingCV || menuState is MenuState.SelectingText) {
                        if (event.action == ACTION_DOWN) {
                            cvUseCase.selectRectangle(x = event.x.toInt(), y = event.y.toInt())
                        }
                        return@launch
                    }

                    selectExistingKeyboardButton(event)
                    selectNewKeyboardButton(event)
                    if (openEditKeyboardDialog()) return@launch

                    if (menuState is MenuState.Keyboard && menuState.recordingKeyboard) {
                        val letter = currentState.keyboard.buttons.firstOrNull {
                            it.contains(x = event.x.toInt(), y = event.y.toInt())
                        }?.text ?: return@launch

                        if (letter.isEmpty()) return@launch
                        if (event.action == ACTION_UP) menuInteractor.appendTypedLetter(letter)
                    }

                    val screenSizes = videoUseCase.observeScreenSizes().value ?: return@launch
                    val bytesArray = controlStreamer.sendControlData(
                        screenSizes = screenSizes,
                        event = event,
                    )

                    recordBytes(bytesArray)

                } finally {
                    event.recycle()
                }
            }
        }
    }

    private fun openEditKeyboardDialog(): Boolean {
        val menuState = menuInteractor.observeMenuState().value
        val selectedRects = cvUseCase.observeSelectedRectangles().value
        if (
            menuState is MenuState.Keyboard &&
            (menuState.editing || menuState.showCvRectangles) &&
            selectedRects.isNotEmpty()
        ) {
            menuInteractor.onEditKeyboardRectangleSelected()
            return true
        }
        return false
    }

    private fun selectExistingKeyboardButton(event: MotionEvent) {
        val menuState = menuInteractor.observeMenuState().value
        if (
            menuState is MenuState.Keyboard &&
            menuState.editing
        ) {
            val button = currentState.keyboard.buttons.firstOrNull {
                it.contains(x = event.x.toInt(), y = event.y.toInt())
            } ?: return
            menuInteractor.setKeyboardOldKey(button.text)
            cvUseCase.setSelectedRectangle(button.rectangle)
        }
    }

    private fun selectNewKeyboardButton(event: MotionEvent) {
        val menuState = menuInteractor.observeMenuState().value
        if (
            menuState is MenuState.Keyboard &&
            menuState.showCvRectangles
        ) {
            menuInteractor.setKeyboardOldKey("")
            cvUseCase.selectRectangle(x = event.x.toInt(), y = event.y.toInt())
        }
    }

    private fun recordBytes(bytesArray: ByteArray?) {
        val menuState = menuInteractor.observeMenuState().value
        if (
            bytesArray == null ||
            !(menuState is MenuState.Keyboard && menuState.recordingKeyboard ||
                    menuState is MenuState.Recording && menuState.controlRecording)
        ) return

        val elapsedMs = if (startRecordingTime == 0L) {
            startRecordingTime = System.nanoTime()
            0L
        } else {
            (System.nanoTime() - startRecordingTime) / 1_000_000L
        }

        val newEvent = Event(
            time = elapsedMs,
            data = bytesArray,
        )

        _stateFlow.update {
            it.copy(
                record = it.record.copy(
                    events = it.record.events + newEvent
                )
            )
        }
    }

    private fun addParam(param: Parameter) {
        _stateFlow.update {
            it.copy(record = it.record.copy(params = it.record.params + param))
        }
    }

    private suspend fun saveScript() {
        val record = currentState.record
        val success = scripterRepository.saveScript(
            Script(
                name = record.recordName,
                node = record.node,
                params = record.params,
                events = record.events,
                timeout = record.timeout,
            ),
        )
        if (!success) return

        uiCommandsFlow.tryEmit(StreamingUiCommand.ShowScriptSavedSnackbar)
        startRecordingTime = 0L
        menuInteractor.onScriptSaved()
        cvUseCase.nextCvMode(CVMode.NO_CV)
        cvUseCase.clearAllRectangles()
        _stateFlow.update {
            it.copy(record = StreamingState.Record())
        }
    }

    private suspend fun getOrResetKeyboard() {
        menuInteractor.setKeyboardLoadingState(true)
        val keyboardResult = scripterRepository.getKeyboard(
            serial = currentState.serial,
            locale = currentState.record.locale,
        )
        if (keyboardResult is ApiResponse.Success && keyboardResult.data.isNotEmpty()) {
            setupKeyboardRects(keyboardResult.data)
            return
        }

        val resetResult = scripterRepository.resetKeyboard(
            serial = currentState.serial,
            locale = currentState.record.locale,
        )
        if (resetResult is ApiResponse.Success) {
            setupKeyboardRects(resetResult.data)
            return
        }
        showKeyboardError()
    }

    private fun setupKeyboardRects(data: List<RectangleWithText>) {
        val screenSizes = videoUseCase.observeScreenSizes().value ?: return
        val buttons = data.mapNotNull {
            val rectangle = it.rectangle ?: return@mapNotNull null
            it.copy(rectangle = rectangle.adjustToClient(screenSizes))
        }

        _stateFlow.update {
            it.copy(keyboard = it.keyboard.copy(buttons = buttons))
        }
        menuInteractor.setKeyboardLoadingState(false)
    }

    private fun showKeyboardError() {
        menuInteractor.setKeyboardLoadingState(false)
        uiCommandsFlow.tryEmit(StreamingUiCommand.ShowNetworkError)
    }

    suspend fun getStreamingData(): ApiResponse<StreamingData> {
        val result = scripterRepository.startSockets(currentState.serial)
        when (result) {
            is ApiResponse.Success -> {
                val fullServerUri = dataStoreRepository.getServerUrl().toUri()
                _stateFlow.update {
                    it.copy(
                        streamingHost = fullServerUri.host.orEmpty(),
                        streamingData = result.data
                    )
                }
            }

            is ApiResponse.Error -> {
                uiCommandsFlow.tryEmit(StreamingUiCommand.ShowNetworkError)
            }
        }
        return result
    }
}
