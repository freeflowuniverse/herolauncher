// Boardmeeting minutes functionality
document.addEventListener('DOMContentLoaded', function() {
  // Only run if we're on the edit page
  if (document.querySelector('.add-resolution')) {
    const resolutionItems = document.querySelector('.resolution-items');
    const addResolutionButton = document.querySelector('.add-resolution');
    
    if (addResolutionButton) {
      addResolutionButton.addEventListener('click', function() {
        const index = document.querySelectorAll('.resolution-item').length;
        
        // Create the main container
        const newResolution = document.createElement('div');
        newResolution.className = 'resolution-item';
        
        // Create title section
        const titleDiv = document.createElement('div');
        const titleLabel = document.createElement('label');
        titleLabel.setAttribute('for', 'resolution_title_' + index);
        titleLabel.textContent = 'Resolution Title';
        const titleInput = document.createElement('input');
        titleInput.setAttribute('type', 'text');
        titleInput.setAttribute('id', 'resolution_title_' + index);
        titleInput.setAttribute('name', 'resolution_titles[]');
        titleInput.setAttribute('placeholder', 'Resolution title');
        titleDiv.appendChild(titleLabel);
        titleDiv.appendChild(titleInput);
        
        // Create text section
        const textDiv = document.createElement('div');
        const textLabel = document.createElement('label');
        textLabel.setAttribute('for', 'resolution_text_' + index);
        textLabel.textContent = 'Resolution Text';
        const textArea = document.createElement('textarea');
        textArea.setAttribute('id', 'resolution_text_' + index);
        textArea.setAttribute('name', 'resolution_texts[]');
        textArea.setAttribute('rows', '3');
        textArea.setAttribute('placeholder', 'Full text of the resolution');
        textDiv.appendChild(textLabel);
        textDiv.appendChild(textArea);
        
        // Create votes grid
        const gridDiv = document.createElement('div');
        gridDiv.className = 'grid';
        
        // Votes For
        const votesForDiv = document.createElement('div');
        const votesForLabel = document.createElement('label');
        votesForLabel.setAttribute('for', 'resolution_votes_for_' + index);
        votesForLabel.textContent = 'Votes For';
        const votesForInput = document.createElement('input');
        votesForInput.setAttribute('type', 'number');
        votesForInput.setAttribute('id', 'resolution_votes_for_' + index);
        votesForInput.setAttribute('name', 'resolution_votes_for[]');
        votesForInput.setAttribute('min', '0');
        votesForInput.setAttribute('value', '0');
        votesForDiv.appendChild(votesForLabel);
        votesForDiv.appendChild(votesForInput);
        
        // Votes Against
        const votesAgainstDiv = document.createElement('div');
        const votesAgainstLabel = document.createElement('label');
        votesAgainstLabel.setAttribute('for', 'resolution_votes_against_' + index);
        votesAgainstLabel.textContent = 'Votes Against';
        const votesAgainstInput = document.createElement('input');
        votesAgainstInput.setAttribute('type', 'number');
        votesAgainstInput.setAttribute('id', 'resolution_votes_against_' + index);
        votesAgainstInput.setAttribute('name', 'resolution_votes_against[]');
        votesAgainstInput.setAttribute('min', '0');
        votesAgainstInput.setAttribute('value', '0');
        votesAgainstDiv.appendChild(votesAgainstLabel);
        votesAgainstDiv.appendChild(votesAgainstInput);
        
        // Votes Abstained
        const votesAbstainedDiv = document.createElement('div');
        const votesAbstainedLabel = document.createElement('label');
        votesAbstainedLabel.setAttribute('for', 'resolution_votes_abstained_' + index);
        votesAbstainedLabel.textContent = 'Abstained';
        const votesAbstainedInput = document.createElement('input');
        votesAbstainedInput.setAttribute('type', 'number');
        votesAbstainedInput.setAttribute('id', 'resolution_votes_abstained_' + index);
        votesAbstainedInput.setAttribute('name', 'resolution_votes_abstained[]');
        votesAbstainedInput.setAttribute('min', '0');
        votesAbstainedInput.setAttribute('value', '0');
        votesAbstainedDiv.appendChild(votesAbstainedLabel);
        votesAbstainedDiv.appendChild(votesAbstainedInput);
        
        // Add all votes to grid
        gridDiv.appendChild(votesForDiv);
        gridDiv.appendChild(votesAgainstDiv);
        gridDiv.appendChild(votesAbstainedDiv);
        
        // Create result section
        const resultDiv = document.createElement('div');
        const resultLabel = document.createElement('label');
        resultLabel.setAttribute('for', 'resolution_result_' + index);
        resultLabel.textContent = 'Result';
        const resultSelect = document.createElement('select');
        resultSelect.setAttribute('id', 'resolution_result_' + index);
        resultSelect.setAttribute('name', 'resolution_results[]');
        
        // Add options to select
        const options = [
          { value: '', text: 'Select result' },
          { value: 'Approved', text: 'Approved' },
          { value: 'Rejected', text: 'Rejected' },
          { value: 'Deferred', text: 'Deferred' },
          { value: 'Withdrawn', text: 'Withdrawn' }
        ];
        
        options.forEach(function(opt) {
          const option = document.createElement('option');
          option.value = opt.value;
          option.textContent = opt.text;
          resultSelect.appendChild(option);
        });
        
        resultDiv.appendChild(resultLabel);
        resultDiv.appendChild(resultSelect);
        
        // Create remove button
        const removeButton = document.createElement('button');
        removeButton.className = 'remove-resolution';
        removeButton.setAttribute('type', 'button');
        removeButton.textContent = 'Remove';
        
        // Add all elements to the resolution item
        newResolution.appendChild(titleDiv);
        newResolution.appendChild(textDiv);
        newResolution.appendChild(gridDiv);
        newResolution.appendChild(resultDiv);
        newResolution.appendChild(removeButton);
        
        // Add to the DOM
        resolutionItems.appendChild(newResolution);
        
        // Add event listener to the new remove button
        removeButton.addEventListener('click', function() {
          resolutionItems.removeChild(newResolution);
        });
      });
    }
    
    // Add event listeners to existing remove buttons
    document.querySelectorAll('.remove-resolution').forEach(button => {
      button.addEventListener('click', function() {
        const resolutionItem = this.closest('.resolution-item');
        resolutionItems.removeChild(resolutionItem);
      });
    });
  }
});
