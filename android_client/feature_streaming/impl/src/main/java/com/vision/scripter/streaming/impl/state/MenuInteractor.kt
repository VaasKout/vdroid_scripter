package com.vision.scripter.streaming.impl.state

import dagger.hilt.android.scopes.ViewModelScoped
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import javax.inject.Inject

@ViewModelScoped
class MenuInteractor @Inject constructor() {

    private val _menuState = MutableStateFlow<MenuState>(MenuState.Usual())
    fun observeMenuState(): StateFlow<MenuState> = _menuState.asStateFlow()

    private val _dialogState = MutableStateFlow(DialogState.NONE)
    fun observeDialogState(): StateFlow<DialogState> = _dialogState.asStateFlow()

    fun onScriptModeClicked() {
        _dialogState.update { DialogState.RECORD }
    }

    fun onExpandClicked() {
        val state = _menuState.value
        if (state is MenuState.Usual) {
            _menuState.update { state.copy(expanded = !state.expanded) }
        }
    }

    fun onRecordingClicked() {
        when (val state = _menuState.value) {
            is MenuState.Recording -> _menuState.update {
                state.copy(controlRecording = !state.controlRecording)
            }

            is MenuState.TypingText -> _menuState.update {
                state.copy(
                    recordingKeyboard = !state.recordingKeyboard,
                    typeText = "",
                )
            }

            else -> Unit
        }
    }

    fun onKeyboardClicked() {
        val state = _menuState.value
        if (state is MenuState.Usual && state.showKeyboardButtons) {
            _menuState.update {
                state.copy(
                    showKeyboardButtons = false,
                    keyboardLoading = false,
                )
            }
            return
        }
        _dialogState.update { DialogState.KEYBOARD }
    }

    fun onCvModeClicked(templateSelectMode: CvSelectMode): CvModeAction? {
        return when (val state = _menuState.value) {
            is MenuState.Recording -> {
                val selectMode = templateSelectMode.incrementOnlyActive()
                _menuState.update { MenuState.SelectingCV(selectMode = selectMode) }
                CvModeAction(newCvMode = CVMode.CV_RECTS)
            }

            is MenuState.SelectingCV -> {
                val selectMode = state.selectMode.increment()
                val cvMode = if (selectMode != CvSelectMode.NONE) {
                    CVMode.CV_RECTS
                } else {
                    CVMode.NO_CV
                }
                _menuState.update { state.copy(selectMode = selectMode) }
                CvModeAction(
                    newCvMode = cvMode,
                    disableSelection = selectMode == CvSelectMode.NONE,
                )
            }

            is MenuState.Usual -> {
                val newCVMode = state.cvMode.increment()
                _menuState.update { state.copy(cvMode = newCVMode) }
                CvModeAction(newCvMode = newCVMode)
            }

            else -> null
        }
    }

    fun onTextModeClicked(): TextModeAction {
        return when (val state = _menuState.value) {
            is MenuState.Recording -> {
                _dialogState.update { DialogState.TEXT }
                TextModeAction.None
            }

            is MenuState.SelectingText -> {
                val newTextMode = state.selectMode.increment()
                _menuState.update { state.copy(selectMode = newTextMode) }
                if (newTextMode != CvSelectMode.NONE) {
                    TextModeAction.SelectAll
                } else {
                    TextModeAction.DisableSelection
                }
            }

            else -> TextModeAction.None
        }
    }

    fun onTextSearchStarted() {
        _dialogState.update { DialogState.NONE }
    }

    fun onTextSearchSuccess(text: String) {
        val state = _menuState.value
        if (state is MenuState.Recording) {
            _menuState.update {
                MenuState.SelectingText(
                    selectMode = state.textSelectMode.increment(),
                    text = text,
                )
            }
        }
    }

    fun onSaveClicked(): SaveAction {
        return when (val state = _menuState.value) {
            is MenuState.SelectingCV -> {
                _menuState.update { MenuState.Recording(templateSelectMode = state.selectMode) }
                SaveAction.SaveTemplate(selectMode = state.selectMode)
            }

            is MenuState.SelectingText -> {
                _menuState.update { MenuState.Recording(textSelectMode = state.selectMode) }
                SaveAction.SaveTextSelection(
                    text = state.text,
                    selectMode = state.selectMode,
                )
            }

            is MenuState.TypingText -> {
                _menuState.update { MenuState.Recording(typeText = state.typeText.isNotEmpty()) }
                SaveAction.SaveTyping(text = state.typeText)
            }

            else -> SaveAction.SaveStep
        }
    }

    fun onStepSaved() {
        _menuState.update { MenuState.Recording() }
    }

    fun onCancelClicked(record: StreamingState.Record): Boolean {
        val wasRecording = _menuState.value is MenuState.Recording
        if (wasRecording) {
            _menuState.update { MenuState.Usual(expanded = true) }
        } else {
            _menuState.update {
                MenuState.Recording(
                    templateSelectMode = record.templateSelectMode,
                    textSelectMode = record.textSelectMode,
                    typeText = record.typeText,
                )
            }
        }
        return wasRecording
    }

    fun onSavedRecordName() {
        _dialogState.update { DialogState.NONE }
        _menuState.update { MenuState.Recording() }
    }

    fun onSaveLocale() {
        _dialogState.update { DialogState.NONE }
        val updated = when (val state = _menuState.value) {
            is MenuState.Recording -> MenuState.TypingText()
            is MenuState.Usual -> state.copy(keyboardLoading = true)
            else -> state
        }
        _menuState.update { updated }
    }

    fun onDialogDismissed(record: StreamingState.Record) {
        val state = _menuState.value
        _dialogState.update { DialogState.NONE }
        if (state is MenuState.Recording) {
            _menuState.update {
                MenuState.Recording(
                    templateSelectMode = record.templateSelectMode,
                    textSelectMode = record.textSelectMode,
                    typeText = record.typeText,
                )
            }
        }
    }

    fun onKeyboardInitStarted() {
        val updated = when (val state = _menuState.value) {
            is MenuState.Usual -> state.copy(keyboardLoading = true)
            is MenuState.TypingText -> state.copy(isLoadingKeyboard = true)
            else -> state
        }
        _menuState.update { updated }
    }

    fun onKeyboardLoaded(hasButtons: Boolean) {
        val updated = when (val state = _menuState.value) {
            is MenuState.Usual -> state.copy(
                keyboardLoading = false,
                showKeyboardButtons = hasButtons,
            )

            is MenuState.TypingText -> state.copy(isLoadingKeyboard = false)
            else -> state
        }
        _menuState.update { updated }
    }

    fun onKeyboardError() {
        val updated = when (val state = _menuState.value) {
            is MenuState.Usual -> state.copy(keyboardLoading = false)
            is MenuState.TypingText -> state.copy(isLoadingKeyboard = false)
            else -> state
        }
        _menuState.update { updated }
    }

    fun appendTypedLetter(letter: String) {
        val state = _menuState.value
        if (state !is MenuState.TypingText) return
        val updatedTypeText = buildString {
            append(state.typeText)
            if (letter == SPACE_KEY) append(" ")
            else append(letter)
        }
        _menuState.update { state.copy(typeText = updatedTypeText) }
    }
}

data class CvModeAction(
    val newCvMode: CVMode,
    val disableSelection: Boolean = false,
)

sealed interface TextModeAction {
    data object None : TextModeAction
    data object SelectAll : TextModeAction
    data object DisableSelection : TextModeAction
}

sealed interface SaveAction {
    data class SaveTemplate(val selectMode: CvSelectMode) : SaveAction
    data class SaveTextSelection(val text: String, val selectMode: CvSelectMode) : SaveAction
    data class SaveTyping(val text: String) : SaveAction
    data object SaveStep : SaveAction
}

enum class DialogState {
    NONE,
    RECORD,
    TEXT,
    KEYBOARD;
}
