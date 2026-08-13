package com.vision.scripter.streaming.impl.blocks.menu.state

import com.vision.scripter.coroutines.api.CoroutineScopeFactory
import com.vision.scripter.data.api.models.adjustToServer
import com.vision.scripter.streaming.impl.blocks.menu.ui.MenuUiCommand
import com.vision.scripter.streaming.impl.blocks.menu.ui.MenuUiState
import com.vision.scripter.streaming.impl.blocks.menu.ui.MenuUiStateHolder
import com.vision.scripter.streaming.impl.data.CvStreamerRepository
import com.vision.scripter.streaming.impl.data.ItemType
import com.vision.scripter.streaming.impl.data.KeyboardRepository
import com.vision.scripter.streaming.impl.data.RecordRepository
import com.vision.scripter.streaming.impl.data.VideoStreamerRepository
import com.vision.scripter.streaming.impl.screen.StreamingEvent
import com.vision.scripter.streaming.impl.screen.StreamingEventsHolder
import com.vision.scripter.streaming.impl.screen.state.CVMode
import com.vision.scripter.streaming.impl.screen.state.KeyboardMode
import com.vision.scripter.streaming.impl.screen.state.increment
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
                if (type !is MenuType.CustomAction) return@update it
                it.copy(type = type.copy(recording = record.recording))
            }
        }.launchIn(coroutineScope)

        keyboardRepository.observeSelectedButton().onEach { oldKey ->
            onEditKeyboardRectangleSelected(oldKey)
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
            _menuState.update {
                it.copy(type = MenuType.SelectingCV(localCvMode = CVMode.CV_RECTS))
            }
            switchCvMode(CVMode.CV_RECTS)
            return
        }

        if (itemType == ItemType.ACTION) {
            _menuState.update { it.copy(type = MenuType.CustomAction()) }
        }
    }

    override fun onKeyboardClicked() {
        _dialogState.update { DialogState.Keyboard }
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

    override fun onCvModeClicked() {
        when (val type = _menuState.value.type) {
            is MenuType.SelectingCV -> {
                val cvMode = type.localCvMode.toggleDetection()
                _menuState.update { it.copy(type = type.copy(localCvMode = cvMode)) }
                switchCvMode(cvMode)
            }

            is MenuType.Usual -> {
                val newCvMode = type.localCvMode.increment()
                _menuState.update { it.copy(type = type.copy(localCvMode = newCvMode)) }
                switchCvMode(newCvMode)
            }

            else -> {}
        }
    }

    override fun onTextModeClicked() {
        val type = _menuState.value.type
        if (type !is MenuType.Usual) return
        if (type.textHighlighted) {
            _menuState.update { it.copy(type = type.copy(textHighlighted = false)) }
            switchCvMode(CVMode.NO_CV)
            return
        }
        _dialogState.update { DialogState.Text }
    }

    override fun onTryToFindText(text: String, locale: String) {
        hideDialog()
        findText(text = text.trim(), locale = locale)
        val type = _menuState.value.type
        if (type is MenuType.Usual) {
            _menuState.update {
                it.copy(
                    type = type.copy(
                        localCvMode = CVMode.NO_CV,
                        textHighlighted = true
                    )
                )
            }
        }
    }

    override fun onKeyboardModeClicked() {
        val type = _menuState.value.type
        if (type !is MenuType.Keyboard) return

        val newMode = type.mode.increment()
        _menuState.update { it.copy(type = type.copy(mode = newMode)) }
        changeKeyboardState(newMode)
    }

    override fun onEditKeyboardButtonSaved(oldKey: String, newKey: String) {
        val type = _menuState.value.type
        if (type !is MenuType.Keyboard) return
        hideDialog()
        editKeyboardKey(oldKey = oldKey, newKey = newKey)
    }

    override fun onSaveClicked() {
        when (_menuState.value.type) {
            is MenuType.Keyboard -> exitKeyboard()
            is MenuType.SelectingCV -> saveImage()
            is MenuType.CustomAction -> saveAction()
            else -> Unit
        }
    }

    override fun onCancelClicked() {
        when (_menuState.value.type) {
            is MenuType.Keyboard -> exitKeyboard()
            else -> _menuState.update { it.copy(type = MenuType.Usual(expanded = true)) }
        }

        coroutineScope.launch {
            recordRepository.clear()
            dropState()
        }
    }

    override fun onExitClicked() {
        uiCommandsFlow.tryEmit(MenuUiCommand.ExitCommand)
    }

    override fun onSaveLocale(locale: String) {
        hideDialog()
        if (_menuState.value.type !is MenuType.Usual) return
        _menuState.update { it.copy(type = MenuType.Keyboard(mode = KeyboardMode.EDIT)) }
        openKeyboard(locale)
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
            notifySaved(saved)
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
            notifySaved(saved)
            recordRepository.clear()
            dropState()
        }
    }

    private fun notifySaved(saved: Boolean) {
        if (saved) {
            eventRepository.sendEvent(StreamingEvent.ShowItemSavedSnackbar)
            return
        }
        eventRepository.sendEvent(StreamingEvent.ShowNetworkError)
    }

    private fun switchCvMode(cvMode: CVMode) {
        coroutineScope.launch {
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
            }
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

    private suspend fun dropState() {
        cvRepository.clearSelectedRectangles()
        cvRepository.nextCvMode(CVMode.NO_CV)
        keyboardRepository.clear()
    }

    private fun setKeyboardLoading(isLoading: Boolean) {
        val type = _menuState.value.type
        if (type is MenuType.Keyboard) {
            _menuState.update { it.copy(type = type.copy(isLoading = isLoading)) }
        }
    }

    private fun exitKeyboard() {
        _menuState.update { it.copy(type = MenuType.Usual(expanded = true)) }
        coroutineScope.launch {
            dropState()
        }
    }

    private fun onEditKeyboardRectangleSelected(oldKey: String) {
        val type = _menuState.value.type
        if (type is MenuType.Keyboard) {
            _dialogState.update { DialogState.EditKeyboard(oldKey) }
        }
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
