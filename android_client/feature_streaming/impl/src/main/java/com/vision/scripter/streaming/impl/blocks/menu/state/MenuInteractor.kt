package com.vision.scripter.streaming.impl.blocks.menu.state

import com.vision.scripter.coroutines.api.CoroutineScopeFactory
import com.vision.scripter.streaming.impl.blocks.menu.ui.MenuUiCommand
import com.vision.scripter.streaming.impl.blocks.menu.ui.MenuUiState
import com.vision.scripter.streaming.impl.blocks.menu.ui.MenuUiStateHolder
import com.vision.scripter.streaming.impl.screen.main.state.CVMode
import com.vision.scripter.streaming.impl.screen.main.state.increment
import com.vision.scripter.streaming.impl.screen.main.state.toggleDetection
import com.vision.scripter.ui.CommandFlow
import dagger.hilt.android.scopes.ViewModelScoped
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.cancel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharingStarted
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.combine
import kotlinx.coroutines.flow.stateIn
import kotlinx.coroutines.flow.update
import javax.inject.Inject

@ViewModelScoped
class MenuInteractor @Inject constructor(
    coroutineScopeFactory: CoroutineScopeFactory,
    uiStateMapper: MenuUiStateMapper,
) : MenuUiStateHolder {

    private val coroutineScope: CoroutineScope =
        coroutineScopeFactory.createBackgroundScope("menu_interactor")

    override val uiCommandsFlow: CommandFlow<MenuUiCommand> =
        CommandFlow(coroutineScope)

    private val _menuState = MutableStateFlow<MenuState>(MenuState.Usual())
    private val menuState = _menuState.asStateFlow()

    private val _dialogState = MutableStateFlow<DialogState>(DialogState.None)
    private val dialogState = _dialogState.asStateFlow()

    override val uiStateFlow: StateFlow<MenuUiState> = combine(
        menuState,
        dialogState,
    ) { menuState, dialogState ->
        uiStateMapper.map(
            state = menuState,
            dialogState = dialogState,
        )
    }.stateIn(coroutineScope, SharingStarted.Eagerly, MenuUiState())

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
        val state = _menuState.value
        if (state is MenuState.Usual) {
            _menuState.update { state.copy(expanded = !state.expanded) }
        }
    }

    override fun onRecordingClicked() {
        when (val state = _menuState.value) {
            is MenuState.Recording -> _menuState.update {
                state.copy(controlRecording = !state.controlRecording)
            }

            is MenuState.Keyboard -> _menuState.update {
                state.copy(
                    recordingKeyboard = !state.recordingKeyboard,
                )
            }

            else -> {}
        }
    }

    override fun onCvModeClicked() {
        when (val state = _menuState.value) {
            is MenuState.Recording -> {
                _menuState.update { MenuState.SelectingCV(cvMode = CVMode.CV_RECTS) }
                uiCommandsFlow.tryEmit(MenuUiCommand.NextCvMode(CVMode.CV_RECTS))
            }

            is MenuState.SelectingCV -> {
                val cvMode = state.cvMode.toggleDetection()
                _menuState.update { state.copy(cvMode = cvMode) }
                uiCommandsFlow.tryEmit(MenuUiCommand.NextCvMode(cvMode))
            }

            is MenuState.Usual -> {
                val newCvMode = state.localCvMode.increment()
                _menuState.update { state.copy(localCvMode = newCvMode) }
                uiCommandsFlow.tryEmit(MenuUiCommand.NextCvMode(newCvMode))
            }

            else -> {}
        }
    }

    override fun onTextModeClicked() {
        val state = _menuState.value
        if (state is MenuState.Usual && state.textHighlighted) {
            _menuState.update { state.copy(textHighlighted = false) }
            uiCommandsFlow.tryEmit(MenuUiCommand.NextCvMode(CVMode.NO_CV))
            return
        }
        _dialogState.update { DialogState.Text }
    }

    override fun onTryToFindText(text: String, locale: String) {
        hideDialog()
        uiCommandsFlow.tryEmit(MenuUiCommand.FindText(text = text.trim(), locale = locale))
    }

    fun onTextSearchSuccess(text: String, locale: String) {
        when (val state = _menuState.value) {
            is MenuState.Recording -> {
                _menuState.update {
                    MenuState.SelectingText
                }
            }

            is MenuState.Usual -> _menuState.update {
                state.copy(localCvMode = CVMode.NO_CV, textHighlighted = true)
            }

            else -> {}
        }
    }

    override fun onKeyboardInitClicked() {
        val state = _menuState.value
        if (state is MenuState.Keyboard) {
            _menuState.update { state.copy(isLoadingKeyboard = true) }
        }
        uiCommandsFlow.tryEmit(MenuUiCommand.KeyboardInit)
    }

    override fun onKeyboardEdited(addNew: Boolean) {
        val state = _menuState.value
        if (state !is MenuState.Keyboard) return
        val newState = when {
            !addNew -> state.copy(editing = !state.editing)
            else -> state.copy(showCvRectangles = !state.showCvRectangles)
        }
        _menuState.update { newState }
    }

    override fun onEditKeyboardButtonSaved(newKey: String) {
        val state = _menuState.value
        if (state !is MenuState.Keyboard) return
        hideDialog()
        uiCommandsFlow.tryEmit(MenuUiCommand.EditKeyboardButton(key = newKey))
    }

    override fun onSaveClicked() {
        uiCommandsFlow.tryEmit(MenuUiCommand.SaveParameter)
        _menuState.update { MenuState.Recording() }
    }

    override fun onCancelClicked() {
        when (val state = _menuState.value) {
            is MenuState.Recording -> {
                _menuState.update { MenuState.Usual(expanded = true) }
                uiCommandsFlow.tryEmit(MenuUiCommand.CancelRecording)
            }

            is MenuState.Keyboard -> {
                if (state.fromUsual) {
                    _menuState.update { MenuState.Usual(expanded = true) }
                    return
                }
                _menuState.update { MenuState.Recording() }
            }

            else -> {
                _menuState.update { MenuState.Recording() }
            }
        }
        uiCommandsFlow.tryEmit(MenuUiCommand.NextCvMode(CVMode.NO_CV))
    }

    override fun onExitClicked() {
        uiCommandsFlow.tryEmit(MenuUiCommand.ExitCommand)
    }

    override fun onTimeoutSaved(timeout: Int) {
        hideDialog()
        val menuState = _menuState.value
        if (menuState is MenuState.Recording) {
            _menuState.update { menuState.copy(recordTimeout = timeout) }
        }
        uiCommandsFlow.tryEmit(MenuUiCommand.SaveTimeout(timeout))
    }

    override fun onSavedRecordName(name: String) {
        hideDialog()
        _menuState.update { MenuState.Recording() }
        uiCommandsFlow.tryEmit(MenuUiCommand.SaveRecordName(name))
    }

    override fun onSaveLocale(locale: String) {
        hideDialog()
        val updated = when (_menuState.value) {
            is MenuState.Recording -> MenuState.Keyboard()
            is MenuState.Usual -> MenuState.Keyboard(fromUsual = true)
            else -> _menuState.value
        }
        _menuState.update { updated }
        uiCommandsFlow.tryEmit(MenuUiCommand.SaveLocale(locale))
    }

    override fun onDialogDismissed() {
        hideDialog()
    }

    fun onEditKeyboardRectangleSelected(oldKey: String) {
        val state = _menuState.value
        if (state is MenuState.Keyboard && (state.editing || state.showCvRectangles)) {
            _dialogState.update { DialogState.EditKeyboard(oldKey) }
        }
    }

    private fun hideDialog() {
        _dialogState.update { DialogState.None }
    }

    fun clear() {
        _menuState.update { MenuState.Usual() }
        _dialogState.update { DialogState.None }
        coroutineScope.cancel()
    }
}
