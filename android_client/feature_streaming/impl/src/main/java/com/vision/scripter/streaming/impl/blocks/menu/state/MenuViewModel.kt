package com.vision.scripter.streaming.impl.blocks.menu.state

import androidx.lifecycle.ViewModel
import dagger.hilt.android.lifecycle.HiltViewModel
import javax.inject.Inject

@HiltViewModel
class MenuViewModel @Inject constructor(
    private val menuInteractor: MenuInteractor,
) : ViewModel(), MenuUiStateHolder by menuInteractor