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
import com.vision.scripter.streaming.impl.screen.main.state.SPACE_KEY
import com.vision.scripter.streaming.impl.screen.main.state.TEMPLATE
import com.vision.scripter.streaming.impl.screen.main.state.TEXT
import com.vision.scripter.streaming.impl.screen.main.state.TYPE_TEXT
import com.vision.scripter.streaming.impl.screen.main.state.YOLO_CLASS
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

    @Volatile
    private var streamJob: Job? = null

    private var startRecordingTime = 0L

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
            is ScreenToVideo.StartLoading -> onLoadData()
            is ScreenToVideo.InitArgs -> initArgs(event.serial)
        }
    }

    private fun handleMenuToVideo(event: MenuToVideo) {
        when (event) {
            is MenuToVideo.SaveClicked -> saveClicked(deriveSaveParameter())
            is MenuToVideo.NextCvMode -> nextCvMode(event.cvMode)
            is MenuToVideo.RecordCancelled -> recordCancelled()
            is MenuToVideo.TextFound -> findTextClicked(event.text, event.locale)
            is MenuToVideo.KeyboardInited -> initKeyboardClicked()
            is MenuToVideo.KeyboardButtonEdited -> editKeyboardKeyClicked(event.newKey)
            is MenuToVideo.LocaleSaved -> saveKeyboardLocaleClicked(event.locale)
            is MenuToVideo.RecordNameSaved -> saveRecordNameClicked(event.name)
            is MenuToVideo.TimeoutSaved -> saveTimeoutClicked(event.timeout)
        }
    }

    private fun deriveSaveParameter(): Parameter? = when (val action = currentState.actionState) {
        is VideoState.ActionState.KeyboardRecording ->
            Parameter(type = TYPE_TEXT, value = action.typeText)

        is VideoState.ActionState.SelectingText -> Parameter(type = TEXT)
        is VideoState.ActionState.SelectingCV -> Parameter(type = TEMPLATE)
        else -> null
    }

    private fun initArgs(serial: String) {
        _stateFlow.update {
            it.copy(serial = serial)
        }
    }

    private fun onLoadData() {
        coroutineScope.launch {
            when (val result = scripterRepository.startSockets(currentState.serial)) {
                is ApiResponse.Success -> {
                    val fullServerUri = dataStoreRepository.getServerUrl().toUri()
                    _stateFlow.update {
                        it.copy(
                            streamingHost = fullServerUri.host.orEmpty(),
                            streamingData = result.data
                        )
                    }

                    sharedEvents.emit(VideoToScreen.SuccessLoading)
                }

                is ApiResponse.Error -> {
                    sharedEvents.emit(VideoToScreen.ShowNetworkError)
                }
            }
        }
        startReactiveStreams()
    }

    private fun nextCvMode(cvMode: CVMode) {
        coroutineScope.launch {
            if (cvMode == CVMode.NO_CV) {
                cvUseCase.restoreSelectedRectangles()
            } else {
                cvUseCase.snapshotSelectedRectangles()
            }
            cvUseCase.nextCvMode(cvMode)
        }
    }

    private fun saveClicked(parameter: Parameter?) {
        coroutineScope.launch {
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
                addParam(Parameter(type = TEMPLATE, value = templateName))
                return@launch
            }
            if (parameter.type == YOLO_CLASS) {
                val label =
                    cvUseCase.observeSelectedRectangles().value.firstOrNull()?.label.orEmpty()
                addParam(Parameter(type = YOLO_CLASS, value = label))
                return@launch
            }
            if (parameter.type.isNotEmpty()) {
                addParam(parameter)
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

            val screenSizes = currentState.screenSizes
            val connectionEstablished =
                videoConnected && controlConnected && cvConnected && screenSizes != null

            _stateFlow.update {
                it.copy(connectionEstablished = connectionEstablished)
            }

            if (!connectionEstablished) {
                _stateFlow.update {
                    it.copy(streamingData = null)
                }
                sharedEvents.emit(VideoToScreen.ShowNetworkError)
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

    override fun onTouchEvent(
        viewWidth: Int,
        viewHeight: Int,
        event: MotionEvent?,
    ) {
        if (event == null) return
        coroutineScope.launch {
            mutex.withLock {
                try {
                    val actionState = currentState.actionState
                    if (
                        actionState == VideoState.ActionState.SelectingCV ||
                        actionState == VideoState.ActionState.SelectingText
                    ) {
                        if (event.action == ACTION_DOWN) {
                            cvUseCase.selectRectangle(x = event.x.toInt(), y = event.y.toInt())
                        }
                        return@launch
                    }

                    if (selectKeyboardKey(event)) return@launch

                    if (actionState is VideoState.ActionState.KeyboardRecording) {
                        val letter = currentState.keyboardButtons.firstOrNull {
                            it.contains(x = event.x.toInt(), y = event.y.toInt())
                        }?.text ?: return@launch

                        if (letter.isEmpty()) return@launch
                        if (event.action == ACTION_UP) appendTypedLetter(letter)
                    }

                    val screenSizes = currentState.screenSizes ?: return@launch
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

    private fun recordCancelled() {
        _stateFlow.update { it.copy(record = VideoState.Record()) }
        cvUseCase.clearAllRectangles()
    }

    private fun findTextClicked(text: String, locale: String) {
        coroutineScope.launch {
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
            sharedEvents.emit(VideoToMenu.TextFound)
        }
    }

    private fun initKeyboardClicked() {
        coroutineScope.launch {
            val result = scripterRepository.resetKeyboard(
                serial = currentState.serial,
                locale = currentState.record.lastParam().locale,
            )
            if (result is ApiResponse.Success) {
                setupKeyboardRects(result.data)
            }
        }
    }

    private fun editKeyboardKeyClicked(key: String) {
        coroutineScope.launch {
            val actionState = (currentState.actionState as? VideoState.ActionState.EditingKeyboard)
                ?: return@launch
            val screenSizes = currentState.screenSizes ?: return@launch
            val success = cvUseCase.editKeyboardSelectedRectangle(
                serial = currentState.serial,
                locale = currentState.record.lastParam().locale,
                oldName = actionState.oldKey,
                newName = key,
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

    private fun saveKeyboardLocaleClicked(locale: String) {
        coroutineScope.launch {
            val newParam = Parameter(type = TYPE_TEXT, locale = locale)
            val updatedParams = currentState.record.params + newParam
            _stateFlow.update {
                it.copy(record = it.record.copy(params = updatedParams))
            }
            getOrResetKeyboard()
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

    private fun selectKeyboardKey(event: MotionEvent): Boolean {
        val actionState = currentState.actionState
        if (actionState is VideoState.ActionState.EditingKeyboard) {
            val button = currentState.keyboardButtons.firstOrNull {
                it.contains(x = event.x.toInt(), y = event.y.toInt())
            } ?: return false
            cvUseCase.setSelectedRectangle(button.rectangle)
            sharedEvents.emit(VideoToMenu.SelectKeyboardKey(button.text))
            return true
        }
        if (actionState == VideoState.ActionState.AddingKeyboardKeys) {
            cvUseCase.selectRectangle(x = event.x.toInt(), y = event.y.toInt())
            sharedEvents.emit(VideoToMenu.SelectKeyboardKey(""))
            return true
        }
        return false
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

        sharedEvents.emit(VideoToScreen.ShowScriptSavedSnackbar)
        startRecordingTime = 0L
        sharedEvents.emit(VideoToMenu.ScriptSaved)

        cvUseCase.nextCvMode(CVMode.NO_CV)
        cvUseCase.clearAllRectangles()
        _stateFlow.update {
            it.copy(record = VideoState.Record())
        }
    }

    private suspend fun getOrResetKeyboard() {
        sharedEvents.emit(VideoToMenu.SetKeyboardLoading(true))
        val keyboardResult = scripterRepository.getKeyboard(
            serial = currentState.serial,
            locale = currentState.record.lastParam().locale,
        )
        if (keyboardResult is ApiResponse.Success && keyboardResult.data.isNotEmpty()) {
            setupKeyboardRects(keyboardResult.data)
            return
        }

        val resetResult = scripterRepository.resetKeyboard(
            serial = currentState.serial,
            locale = currentState.record.lastParam().locale,
        )
        if (resetResult is ApiResponse.Success) {
            setupKeyboardRects(resetResult.data)
            return
        }
        showKeyboardError()
    }

    private fun setupKeyboardRects(data: List<RectangleWithText>) {
        val screenSizes = currentState.screenSizes ?: return
        val buttons = data.mapNotNull {
            val rectangle = it.rectangle ?: return@mapNotNull null
            it.copy(rectangle = rectangle.adjustToClient(screenSizes))
        }

        _stateFlow.update {
            it.copy(keyboardButtons = buttons)
        }
        sharedEvents.emit(VideoToMenu.SetKeyboardLoading(false))
    }

    fun appendTypedLetter(letter: String) {
        val actionState = currentState.actionState
        if (actionState !is VideoState.ActionState.KeyboardRecording) return

        val updatedTypeText = buildString {
            append(actionState.typeText)
            if (letter == SPACE_KEY) append(" ")
            else append(letter)
        }
        val updatedActionState = actionState.copy(typeText = updatedTypeText)
        _stateFlow.update { currentState.copy(actionState = updatedActionState) }
    }

    private fun showKeyboardError() {
        sharedEvents.emit(VideoToMenu.SetKeyboardLoading(false))
        sharedEvents.emit(VideoToScreen.ShowNetworkError)
    }
}
