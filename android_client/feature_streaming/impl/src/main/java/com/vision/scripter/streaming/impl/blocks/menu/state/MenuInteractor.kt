package com.vision.scripter.streaming.impl.blocks.menu.state

import com.vision.scripter.coroutines.api.CoroutineScopeFactory
import com.vision.scripter.data.api.models.Parameter
import com.vision.scripter.data.api.models.adjustToServer
import com.vision.scripter.streaming.impl.blocks.menu.ui.MenuUiCommand
import com.vision.scripter.streaming.impl.blocks.menu.ui.MenuUiState
import com.vision.scripter.streaming.impl.blocks.menu.ui.MenuUiStateHolder
import com.vision.scripter.streaming.impl.data.CvStreamerRepository
import com.vision.scripter.streaming.impl.data.KeyboardRepository
import com.vision.scripter.streaming.impl.data.RecordRepository
import com.vision.scripter.streaming.impl.data.VideoStreamerRepository
import com.vision.scripter.streaming.impl.screen.StreamingEvent
import com.vision.scripter.streaming.impl.screen.StreamingEventsHolder
import com.vision.scripter.streaming.impl.screen.state.CVMode
import com.vision.scripter.streaming.impl.screen.state.KeyboardMode
import com.vision.scripter.streaming.impl.screen.state.TEMPLATE
import com.vision.scripter.streaming.impl.screen.state.TEXT
import com.vision.scripter.streaming.impl.screen.state.YOLO_CLASS
import com.vision.scripter.streaming.impl.screen.state.increment
import com.vision.scripter.streaming.impl.screen.state.toType
import com.vision.scripter.streaming.impl.screen.state.toggleDetection
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
    private val cvRepository: CvStreamerRepository,
    private val keyboardRepository: KeyboardRepository,
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
                if (type !is MenuType.Recording) return@update it
                it.copy(type = type.copy(controlRecording = record.controlRecording))
            }
        }.launchIn(coroutineScope)

        keyboardRepository.observeSelectedButton().onEach { oldKey ->
            onEditKeyboardRectangleSelected(oldKey)
        }.launchIn(coroutineScope)
    }

    override fun onScriptModeClicked() {
        _dialogState.update { DialogState.Record }
    }

    override fun onTimeoutClicked() {
        _dialogState.update { DialogState.Timeout }
    }

    override fun onKeyboardClicked() {
        _dialogState.update { DialogState.Keyboard }
    }

    override fun onExpandClicked() {
        val type = _menuState.value.type
        if (type is MenuType.Usual) {
            updateMenuType(type.copy(expanded = !type.expanded))
        }
    }

    override fun onRecordingClicked() {
        val state = _menuState.value
        if (state.type !is MenuType.Recording) return
        recordRepository.switchControlRecording()
    }

    override fun onCvModeClicked() {
        when (val type = _menuState.value.type) {
            is MenuType.Recording -> {
                updateMenuType(MenuType.SelectingCV(localCvMode = CVMode.CV_RECTS))
                nextCvMode(cvMode = CVMode.CV_RECTS, recording = true)
            }

            is MenuType.SelectingCV -> {
                val cvMode = type.localCvMode.toggleDetection()
                updateMenuType(type.copy(localCvMode = cvMode))
                nextCvMode(cvMode = cvMode, recording = true)
            }

            is MenuType.Usual -> {
                val newCvMode = type.localCvMode.increment()
                updateMenuType(type.copy(localCvMode = newCvMode))
                nextCvMode(cvMode = newCvMode, recording = false)
            }

            else -> {}
        }
    }

    override fun onTextModeClicked() {
        val type = _menuState.value.type
        if (type is MenuType.Usual && type.textHighlighted) {
            updateMenuType(type.copy(textHighlighted = false))
            nextCvMode(cvMode = CVMode.NO_CV, recording = false)
            return
        }
        _dialogState.update { DialogState.Text }
    }

    override fun onTryToFindText(text: String, locale: String) {
        hideDialog()
        findText(text = text.trim(), locale = locale)
        when (val type = _menuState.value.type) {
            is MenuType.Recording -> {
                updateMenuType(MenuType.SelectingText)
            }

            is MenuType.Usual -> {
                updateMenuType(type.copy(localCvMode = CVMode.NO_CV, textHighlighted = true))
            }

            else -> {}
        }
    }

    override fun onKeyboardModeClicked() {
        val type = _menuState.value.type
        if (type !is MenuType.Keyboard) return

        val newMode = type.mode.increment()
        updateMenuType(type.copy(mode = newMode))
        changeKeyboardState(newMode)
    }

    override fun onEditKeyboardButtonSaved(oldKey: String, newKey: String) {
        val type = _menuState.value.type
        if (type !is MenuType.Keyboard) return
        hideDialog()
        editKeyboardKey(oldKey = oldKey, newKey = newKey)
    }

    override fun onSaveClicked() {
        val type = _menuState.value.type
        when (type) {
            is MenuType.Keyboard -> exitKeyboard(type)

            is MenuType.SelectingText, is MenuType.SelectingCV -> {
                updateMenuType(MenuType.Recording())
            }

            is MenuType.Recording -> {
                updateMenuType(MenuType.Usual(expanded = true))
            }

            else -> Unit
        }

        coroutineScope.launch {
            when (type) {
                is MenuType.Keyboard -> saveTypedText()
                is MenuType.SelectingText, is MenuType.SelectingCV -> saveParameter()
                is MenuType.Recording -> saveScript()
                else -> Unit
            }
        }
    }

    override fun onCancelClicked() {
        when (val type = _menuState.value.type) {
            is MenuType.Recording -> {
                updateMenuType(MenuType.Usual(expanded = true))
            }

            is MenuType.Keyboard -> exitKeyboard(type)

            else -> {
                updateMenuType(MenuType.Recording())
            }
        }

        coroutineScope.launch {
            val tmpParam = recordRepository.observeRecord().value.tmpParam
            recordRepository.clear(saveName = tmpParam != null)
            dropState()
        }
    }

    override fun onExitClicked() {
        uiCommandsFlow.tryEmit(MenuUiCommand.ExitCommand)
    }

    override fun onTimeoutSaved(timeout: Int) {
        hideDialog()
        val type = _menuState.value.type
        if (type is MenuType.Recording) {
            updateMenuType(type.copy(recordTimeout = timeout))
        }
        recordRepository.updateTimeout(timeout)
    }

    override fun onSavedRecordName(name: String) {
        hideDialog()
        updateMenuType(MenuType.Recording())
        recordRepository.updateName(name)
    }

    override fun onSaveLocale(locale: String) {
        hideDialog()
        val updated = when (_menuState.value.type) {
            is MenuType.Recording -> MenuType.Keyboard(mode = KeyboardMode.TYPING)
            is MenuType.Usual -> MenuType.Keyboard(fromUsual = true, mode = KeyboardMode.TYPING)
            else -> return
        }
        updateMenuType(updated)
        openKeyboard(locale)
    }

    override fun onDialogDismissed() {
        hideDialog()
    }

    private fun nextCvMode(cvMode: CVMode, recording: Boolean) {
        coroutineScope.launch {
            if (recording) {
                val type = cvMode.toType()
                if (type.isNotEmpty()) {
                    recordRepository.updateTmpParam(Parameter(type = type))
                }
            }
            if (cvMode == CVMode.NO_CV) {
                cvRepository.restoreSelectedRectangles()
            } else {
                cvRepository.snapshotSelectedRectangles()
            }
            cvRepository.nextCvMode(cvMode)
        }
    }

    private fun findText(text: String, locale: String) {
        coroutineScope.launch {
            val screenSizes = videoRepository.observeScreenSizes().value ?: return@launch
            val found = cvRepository.findTextRectangles(
                serial = serial,
                text = text,
                locale = locale,
                screenSizes = screenSizes,
            )
            if (!found) {
                eventRepository.sendEvent(StreamingEvent.ShowNetworkError)
                return@launch
            }
            recordRepository.updateTmpParam(
                Parameter(type = TEXT, value = text, locale = locale),
            )
        }
    }

    private fun changeKeyboardState(keyboardMode: KeyboardMode) {
        coroutineScope.launch {
            keyboardRepository.updateKeyboardState(keyboardMode)
            cvRepository.clearSelectedRectangles()
            if (keyboardMode == KeyboardMode.ADD_NEW) {
                cvRepository.nextCvMode(CVMode.CV_RECTS)
                return@launch
            }
            cvRepository.nextCvMode(CVMode.NO_CV)
        }
    }

    private fun editKeyboardKey(oldKey: String, newKey: String) {
        coroutineScope.launch {
            val screenSizes = videoRepository.observeScreenSizes().value ?: return@launch
            val tmpZone =
                cvRepository.observeSelectedRectangles().value.firstOrNull() ?: return@launch
            val success = keyboardRepository.editKeyboardKey(
                serial = serial,
                oldName = oldKey,
                newName = newKey,
                rectangle = tmpZone.adjustToServer(screenSizes),
            )
            cvRepository.clearSelectedRectangles()
            if (success) {
                getOrResetKeyboard()
                return@launch
            }
            eventRepository.sendEvent(StreamingEvent.ShowNetworkError)
        }
    }

    private fun openKeyboard(locale: String) {
        coroutineScope.launch {
            keyboardRepository.updateKeyboardLocale(locale)
            getOrResetKeyboard()
        }
    }

    private suspend fun getOrResetKeyboard() {
        val screenSizes = videoRepository.observeScreenSizes().value ?: return
        setKeyboardLoading(true)
        val loaded = keyboardRepository.getOrResetKeyboard(
            serial = serial,
            screenSizes = screenSizes,
        )
        setKeyboardLoading(false)
        if (!loaded) eventRepository.sendEvent(StreamingEvent.ShowNetworkError)
    }

    private suspend fun saveTypedText() {
        val param = keyboardRepository.extractParameter()
        if (param == null) {
            recordRepository.clear()
            dropState()
            return
        }
        addParam(param)
    }

    private suspend fun saveParameter() {
        val record = recordRepository.observeRecord().value
        val parameter = record.tmpParam
        if (parameter == null) {
            saveScript()
            return
        }

        if (parameter.type == TEMPLATE) {
            val screenSizes = videoRepository.observeScreenSizes().value ?: return
            cvRepository.nextCvMode(CVMode.NO_CV)
            val templateName = "${record.params.size + 1}"
            cvRepository.saveSelectedRectangle(
                serial = serial,
                node = record.node,
                name = record.name,
                value = templateName,
                screenSizes = screenSizes,
            )
            addParam(parameter.copy(value = templateName))
            return
        }

        if (parameter.type == YOLO_CLASS) {
            val label =
                cvRepository.observeSelectedRectangles().value.firstOrNull()?.label.orEmpty()
            cvRepository.nextCvMode(CVMode.NO_CV)
            addParam(parameter.copy(value = label))
            return
        }

        if (parameter.value.isNotEmpty()) {
            addParam(parameter)
        }
    }

    private suspend fun saveScript() {
        val success = recordRepository.saveScript()
        if (!success) return

        eventRepository.sendEvent(StreamingEvent.ShowScriptSavedSnackbar)
        recordRepository.clear()
        dropState()
    }

    private suspend fun addParam(param: Parameter) {
        recordRepository.addParam(param)
        dropState()
    }

    private suspend fun dropState() {
        cvRepository.clearSelectedRectangles()
        cvRepository.nextCvMode(CVMode.NO_CV)
        keyboardRepository.clear()
    }

    private fun setKeyboardLoading(isLoading: Boolean) {
        val type = _menuState.value.type
        if (type is MenuType.Keyboard) {
            updateMenuType(type.copy(isLoading = isLoading))
        }
    }

    private fun exitKeyboard(type: MenuType.Keyboard) {
        if (type.fromUsual) {
            updateMenuType(MenuType.Usual(expanded = true))
            return
        }
        updateMenuType(MenuType.Recording())
    }

    private fun onEditKeyboardRectangleSelected(oldKey: String) {
        val type = _menuState.value.type
        if (type is MenuType.Keyboard && type.mode != KeyboardMode.TYPING) {
            _dialogState.update { DialogState.EditKeyboard(oldKey) }
        }
    }

    private fun updateMenuType(type: MenuType) {
        _menuState.update { it.copy(type = type) }
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
