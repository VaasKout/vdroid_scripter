package com.vision.scripter.streaming.impl.blocks.video.state

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
import com.vision.scripter.data.api.models.adjustToClient
import com.vision.scripter.data.api.models.contains
import com.vision.scripter.network.api.ApiResponse
import com.vision.scripter.prefs.api.DataStoreRepository
import com.vision.scripter.streaming.impl.blocks.video.ui.VideoUiState
import com.vision.scripter.streaming.impl.blocks.video.ui.VideoUiStateHolder
import com.vision.scripter.streaming.impl.screen.main.state.CVMode
import com.vision.scripter.streaming.impl.screen.main.state.KeyboardState
import com.vision.scripter.streaming.impl.screen.main.state.SPACE_KEY
import com.vision.scripter.streaming.impl.screen.main.state.TEMPLATE
import com.vision.scripter.streaming.impl.screen.main.state.TEXT
import com.vision.scripter.streaming.impl.screen.main.state.TYPE_TEXT
import com.vision.scripter.streaming.impl.screen.main.state.YOLO_CLASS
import com.vision.scripter.streaming.impl.screen.main.state.toType
import com.vision.scripter.streaming.impl.shared.MenuToVideo
import com.vision.scripter.streaming.impl.shared.ScreenToVideo
import com.vision.scripter.streaming.impl.shared.StreamingSharedEventsHolder
import com.vision.scripter.streaming.impl.shared.VideoToMenu
import com.vision.scripter.streaming.impl.shared.VideoToScreen
import com.vision.scripter.streaming.impl.usecases.CvUseCase
import com.vision.scripter.streaming.impl.usecases.VideoUseCase
import dagger.hilt.android.scopes.ViewModelScoped
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Job
import kotlinx.coroutines.cancel
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
import kotlinx.coroutines.sync.Mutex
import kotlinx.coroutines.sync.withLock
import javax.inject.Inject

@ViewModelScoped
class VideoInteractor @Inject constructor(
    coroutineScopeFactory: CoroutineScopeFactory,
    uiStateMapper: VideoUiStateMapper,
    private val scripterRepository: ScripterRepository,
    private val dataStoreRepository: DataStoreRepository,
    private val videoUseCase: VideoUseCase,
    private val controlStreamer: ControlStreamer,
    private val cvUseCase: CvUseCase,
    private val sharedEvents: StreamingSharedEventsHolder,
) : VideoUiStateHolder {

    private val coroutineScope: CoroutineScope =
        coroutineScopeFactory.createBackgroundScope("video_interactor")

    private val _stateFlow = MutableStateFlow(VideoState())
    private val stateFlow = _stateFlow.asStateFlow()

    private val currentState: VideoState
        get() = _stateFlow.value

    override val uiStateFlow: StateFlow<VideoUiState> = stateFlow
        .map(uiStateMapper::map)
        .stateIn(coroutineScope, SharingStarted.Eagerly, VideoUiState())

    fun clear() {
        coroutineScope.cancel()
    }

    private val mutex = Mutex()
    private val touchMutex = Mutex()

    @Volatile
    private var streamJob: Job? = null

    private var startRecordingTime = 0L

    init {
        startReactiveStreams()
    }

    private fun startReactiveStreams() {
        cvUseCase.observeRectangles().onEach { rectangles ->
            _stateFlow.update {
                it.copy(cvRectangles = rectangles)
            }
        }.launchIn(coroutineScope)

        cvUseCase.observeSelectedRectangles().onEach { selected ->
            _stateFlow.update {
                it.copy(selectedRectangles = selected)
            }
        }.launchIn(coroutineScope)

        videoUseCase.observeScreenSizes().onEach { screenSizes ->
            _stateFlow.update {
                it.copy(screenSizes = screenSizes)
            }
        }.launchIn(coroutineScope)

        sharedEvents.sharedEventsFlow.onEach { event ->
            when (event) {
                is ScreenToVideo -> handleScreenToVideo(event)
                is MenuToVideo -> handleMenuToVideo(event)
                else -> Unit
            }
        }.launchIn(coroutineScope)
    }

    private fun handleScreenToVideo(event: ScreenToVideo) {
        when (event) {
            is ScreenToVideo.StartLoading -> onLoadData(event.serial)
        }
    }

    private fun handleMenuToVideo(event: MenuToVideo) {
        when (event) {
            is MenuToVideo.SaveClicked -> onSaveClicked()
            is MenuToVideo.OnCvModeClicked -> nextCvMode(event.nextMode, event.recording)
            is MenuToVideo.OnTextModeClicked -> nextCvMode(CVMode.NO_CV, false)
            is MenuToVideo.CancelClicked -> onCancelClicked()
            is MenuToVideo.FindText -> findTextClicked(event.text, event.locale)
            is MenuToVideo.KeyboardStateChanged -> changeKeyboardState(event.keyboardState)
            is MenuToVideo.KeyboardButtonEdited -> editKeyboardKey(
                oldKey = event.oldKey,
                newKey = event.newKey,
            )

            is MenuToVideo.KeyboardLocaleSaved -> openKeyboard(
                locale = event.locale,
                mode = event.mode,
            )

            is MenuToVideo.RecordNameSaved -> saveRecordNameClicked(event.name)
            is MenuToVideo.TimeoutSaved -> saveTimeoutClicked(event.timeout)
            is MenuToVideo.OnRecordingClicked -> onRecordingClicked(event.isControlRecording)
        }
    }

    private fun onLoadData(serial: String) {
        coroutineScope.launch {
            _stateFlow.update {
                it.copy(serial = serial)
            }
            when (val result = scripterRepository.startSockets(currentState.serial)) {
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
                    sharedEvents.emit(VideoToScreen.ShowNetworkError)
                }
            }
        }
    }

    private fun nextCvMode(cvMode: CVMode, recording: Boolean) {
        coroutineScope.launch {
            if (recording) {
                val type = cvMode.toType()
                if (type.isNotEmpty()) {
                    _stateFlow.update {
                        it.copy(
                            tmpParam = Parameter(type = type),
                        )
                    }
                }
            }
            if (cvMode == CVMode.NO_CV) {
                cvUseCase.restoreSelectedRectangles()
            } else {
                cvUseCase.snapshotSelectedRectangles()
            }
            cvUseCase.nextCvMode(cvMode)
        }
    }

    private fun onSaveClicked() {
        coroutineScope.launch {
            mutex.withLock {
                if (currentState.keyboard.buttons.isNotEmpty()) {
                    saveKeyboard()
                    return@launch
                }

                val parameter = currentState.tmpParam
                if (parameter == null) {
                    saveScript()
                    return@launch
                }

                if (parameter.type == TEMPLATE) {
                    val screenSizes = currentState.screenSizes ?: return@launch
                    cvUseCase.nextCvMode(CVMode.NO_CV)
                    val templateName = "${currentState.record.params.size + 1}"
                    cvUseCase.saveSelectedRectangle(
                        serial = currentState.serial,
                        node = currentState.record.node,
                        name = currentState.record.recordName,
                        value = templateName,
                        screenSizes = screenSizes,
                    )
                    addParam(parameter.copy(value = templateName))
                    return@launch
                }

                if (parameter.type == YOLO_CLASS) {
                    val label =
                        cvUseCase.observeSelectedRectangles().value.firstOrNull()?.label.orEmpty()
                    cvUseCase.nextCvMode(CVMode.NO_CV)
                    addParam(parameter.copy(value = label))
                    return@launch
                }

                if (parameter.value.isNotEmpty()) {
                    addParam(parameter)
                }
            }
        }
    }

    private fun onRecordingClicked(isControlRecording: Boolean) {
        _stateFlow.update {
            it.copy(record = it.record.copy(controlRecording = isControlRecording))
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

            val screenSizes = currentState.screenSizes
            val connectionEstablished =
                videoConnected && controlConnected && cvConnected && screenSizes != null

            if (!connectionEstablished) {
                _stateFlow.update {
                    it.copy(streamingData = null)
                }
                sharedEvents.emit(VideoToScreen.ShowNetworkError)
                return@launch
            }

            sharedEvents.emit(VideoToScreen.SuccessLoading)

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

    override fun onTouchEvent(
        viewWidth: Int,
        viewHeight: Int,
        event: MotionEvent?,
    ) {
        if (event == null) return
        val screenSizes = currentState.screenSizes ?: return
        coroutineScope.launch {
            touchMutex.withLock {
                try {
                    val tmpParam = currentState.tmpParam
                    if (tmpParam?.type == TEMPLATE || tmpParam?.type == YOLO_CLASS) {
                        if (event.action == ACTION_DOWN) {
                            cvUseCase.selectRectangle(x = event.x.toInt(), y = event.y.toInt())
                        }
                        return@launch
                    }

                    if (currentState.keyboard.buttons.isNotEmpty()) {
                        handleKeyboard(event)
                        return@launch
                    }

                    val bytesArray = controlStreamer.sendControlData(
                        screenSizes = screenSizes,
                        event = event,
                    )

                    if (currentState.record.controlRecording) recordBytes(bytesArray)
                } finally {
                    event.recycle()
                }
            }
        }
    }

    private fun onCancelClicked() {
        coroutineScope.launch {
            mutex.withLock {
                val parameter = currentState.tmpParam
                val newRecordName = if (parameter != null) currentState.record.recordName else ""
                dropState(record = VideoState.Record(recordName = newRecordName))
            }
        }
    }

    private fun findTextClicked(text: String, locale: String) {
        coroutineScope.launch {
            mutex.withLock {
                val screenSizes = currentState.screenSizes ?: return@launch
                val found = cvUseCase.findTextRectangles(
                    serial = currentState.serial,
                    text = text,
                    locale = locale,
                    screenSizes = screenSizes,
                )
                if (!found) {
                    sharedEvents.emit(VideoToScreen.ShowNetworkError)
                    return@launch
                }
                _stateFlow.update {
                    it.copy(tmpParam = Parameter(type = TEXT, value = text, locale = locale))
                }
            }
        }
    }

    private fun changeKeyboardState(keyboardState: KeyboardState) {
        coroutineScope.launch {
            mutex.withLock {
                _stateFlow.update {
                    it.copy(keyboard = it.keyboard.copy(mode = keyboardState))
                }
                cvUseCase.clearSelectedRectangles()

                if (keyboardState == KeyboardState.ADD_NEW) {
                    cvUseCase.nextCvMode(CVMode.CV_RECTS)
                    return@withLock
                }
                cvUseCase.nextCvMode(CVMode.NO_CV)
            }
        }
    }

    private fun editKeyboardKey(oldKey: String, newKey: String) {
        coroutineScope.launch {
            mutex.withLock {
                val locale = currentState.keyboard.locale.ifEmpty { return@launch }
                val screenSizes = currentState.screenSizes ?: return@launch
                val success = cvUseCase.editKeyboardSelectedRectangle(
                    serial = currentState.serial,
                    locale = locale,
                    oldName = oldKey,
                    newName = newKey,
                    screenSizes = screenSizes,
                )
                cvUseCase.clearSelectedRectangles()
                if (success) {
                    getOrResetKeyboard()
                    return@launch
                }
                sharedEvents.emit(VideoToScreen.ShowNetworkError)
            }
        }
    }

    private fun openKeyboard(locale: String, mode: KeyboardState) {
        coroutineScope.launch {
            mutex.withLock {
                _stateFlow.update {
                    it.copy(keyboard = VideoState.Keyboard(locale = locale, mode = mode))
                }
                getOrResetKeyboard()
            }
        }
    }

    private fun saveTimeoutClicked(timeout: Int) {
        _stateFlow.update {
            it.copy(record = it.record.copy(timeout = timeout))
        }
    }

    private fun saveRecordNameClicked(name: String) {
        _stateFlow.update {
            it.copy(record = it.record.copy(recordName = name))
        }
    }

    private fun handleKeyboard(event: MotionEvent) {
        when (currentState.keyboard.mode) {
            KeyboardState.TYPING -> { recordLetters(event)
            }
            KeyboardState.EDIT -> selectKeyboardKey(event)
            KeyboardState.ADD_NEW -> selectNewKeyboardRect(event)
        }
    }

    private fun recordLetters(event: MotionEvent) {
        if (event.action != ACTION_UP) return
        if (currentState.record.recordName.isEmpty()) return

        val keyboard = currentState.keyboard
        if (keyboard.locale.isEmpty()) return
        val letter = keyboard.buttons.firstOrNull {
            it.contains(x = event.x.toInt(), y = event.y.toInt())
        }?.text ?: return
        if (letter.isEmpty()) return

        val updatedText = buildString {
            append(keyboard.typedText)
            if (letter == SPACE_KEY) append(" ")
            else append(letter)
        }
        _stateFlow.update {
            it.copy(keyboard = it.keyboard.copy(typedText = updatedText))
        }
    }

    private fun selectKeyboardKey(event: MotionEvent) {
        if (event.action != ACTION_DOWN) return
        val button = currentState.keyboard.buttons.firstOrNull {
            it.contains(x = event.x.toInt(), y = event.y.toInt())
        } ?: return
        cvUseCase.setSelectedRectangle(button.rectangle)
        sharedEvents.emit(VideoToMenu.SelectKeyboardKey(button.text))
    }

    private fun selectNewKeyboardRect(event: MotionEvent) {
        if (event.action != ACTION_DOWN) return
        cvUseCase.selectRectangle(x = event.x.toInt(), y = event.y.toInt())
        sharedEvents.emit(VideoToMenu.SelectKeyboardKey(""))
    }

    private suspend fun saveKeyboard() {
        val keyboard = currentState.keyboard
        if (keyboard.typedText.isEmpty()) {
            dropState()
            return
        }
        addParam(
            Parameter(
                type = TYPE_TEXT,
                value = keyboard.typedText,
                locale = keyboard.locale,
            ),
        )
    }

    private suspend fun dropState(record: VideoState.Record? = null) {
        _stateFlow.update {
            it.copy(
                record = record ?: VideoState.Record(),
                tmpParam = null,
                keyboard = VideoState.Keyboard(),
            )
        }
        cvUseCase.clearSelectedRectangles()
        cvUseCase.nextCvMode(CVMode.NO_CV)
    }

    private fun recordBytes(bytesArray: ByteArray?) {
        if (bytesArray == null) return

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
        coroutineScope.launch {
            val updatedRecord = currentState.record.copy(
                params = currentState.record.params + param,
            )
            dropState(record = updatedRecord)
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

        sharedEvents.emit(VideoToScreen.ShowScriptSavedSnackbar)
        startRecordingTime = 0L
        dropState()
    }

    private suspend fun getOrResetKeyboard() {
        val locale = currentState.keyboard.locale.ifEmpty { return }
        sharedEvents.emit(VideoToMenu.SetKeyboardLoading(true))
        val keyboardResult = scripterRepository.getKeyboard(
            serial = currentState.serial,
            locale = locale,
        )
        if (keyboardResult is ApiResponse.Success && keyboardResult.data.isNotEmpty()) {
            setupKeyboardRects(keyboardResult.data)
            return
        }

        val resetResult = scripterRepository.resetKeyboard(
            serial = currentState.serial,
            locale = locale,
        )
        if (resetResult is ApiResponse.Success) {
            setupKeyboardRects(resetResult.data)
            return
        }
        sharedEvents.emit(VideoToMenu.SetKeyboardLoading(false))
        sharedEvents.emit(VideoToScreen.ShowNetworkError)
    }

    private fun setupKeyboardRects(data: List<RectangleWithText>) {
        val screenSizes = currentState.screenSizes ?: return
        val buttons = data.mapNotNull {
            val rectangle = it.rectangle ?: return@mapNotNull null
            it.copy(rectangle = rectangle.adjustToClient(screenSizes))
        }

        _stateFlow.update {
            it.copy(keyboard = it.keyboard.copy(buttons = buttons))
        }
        sharedEvents.emit(VideoToMenu.SetKeyboardLoading(false))
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
}
