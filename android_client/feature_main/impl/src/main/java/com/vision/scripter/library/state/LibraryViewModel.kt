package com.vision.scripter.library.state

import androidx.lifecycle.ViewModel
import dagger.hilt.android.lifecycle.HiltViewModel
import javax.inject.Inject

@HiltViewModel
internal class LibraryViewModel @Inject constructor(
    private val libraryInteractor: LibraryInteractor,
) : ViewModel(), LibraryUiStateHolder by libraryInteractor {

    override fun onCleared() {
        super.onCleared()
        libraryInteractor.clear()
    }
}
