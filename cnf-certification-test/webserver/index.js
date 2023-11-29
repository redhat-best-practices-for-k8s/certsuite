let yamlGlobal
var counters = {};
function addCrdElem(counters,name1,name2,clickedId){

    console.log(clickedId)

    if (!counters[clickedId]|| counters[clickedId] <= 1 ) {
                counters[clickedId] = 1;
                var i= counters[clickedId]
                var t1=clickedId+name1+i+''
                var t2=clickedId+name2+i+''
                var field = 
                '<h7 for="'+t1+'">'+name1+':</h7><pf-text-input id='+t1+
                ' name="'+clickedId+name1+'" ></pf-text-input>' + 
                '<h7 for="'+t2+'">'+name2+':</h7><pf-text-input id='+t2+
                ' name="'+clickedId+name2+'" ></pf-text-input>' ;
                var target=clickedId+'add'
                $('#'+target).after(field);
            }
            else{
                var i= counters[clickedId]
                var t1=clickedId+name1+i+''
                var t2=clickedId+name2+i+''
                var field =
                '<h7 for="'+t1+'">'+name1+':</h7><pf-text-input id='+t1+
                ' name="'+clickedId+name1+'" ></pf-text-input>' + 
                '<h7 for="'+t2+'">'+name2+':</h7><pf-text-input id='+t2+
                ' name="'+clickedId+name2+'" ></pf-text-input>' ;
                var target=clickedId+name2+(i-1)+''
                $('#'+target).after(field);
            }
            return counters
          }
 

  $(document).ready(function() {
    selectScenario('lifecycle')
    selectScenario('manageability')
    selectScenario('affiliated-certification')
    selectScenario('operator')
    selectScenario('access-control')
    selectScenario('platform-alteration')
    selectScenario('networking')
    selectScenario('performance')
    selectScenario('observability')
    updateTestCounter('lifecycle')
    updateAllCounter()
    document.getElementById('targetNameSpacesremove').style.display = 'none';
    document.getElementById('podsUnderTestLabelsremove').style.display = 'none';
    document.getElementById('operatorsUnderTestLabelsremove').style.display = 'none';
    document.getElementById('targetCrdFiltersremove').style.display = 'none';
    document.getElementById('managedDeploymentsremove').style.display = 'none';
    document.getElementById('managedStatefulsetsremove').style.display = 'none';
    document.getElementById('acceptedKernelTaintsremove').style.display = 'none';
    document.getElementById('skipHelmChartListremove').style.display = 'none';
    document.getElementById('servicesignorelistremove').style.display = 'none';
    document.getElementById('skipScalingTestDeploymentsremove').style.display = 'none';
    document.getElementById('skipScalingTestStatefulsetsremove').style.display = 'none';
    document.getElementById('ValidProtocolNamesremove').style.display = 'none';
    
    const inputElement = document.getElementById('tnfFile')
    inputElement.addEventListener('change', handleTnfFiles, false)
      $('.add').on('click', function() {
        var clickedIdOrg = $(this).attr('id');
        var clickedId = clickedIdOrg.replace('add', '');
        if (clickedId == 'targetCrdFilters'){
          counters=addCrdElem(counters,'nameSuffix','scalable',clickedId)
        }
        else if (clickedId == 'skipScalingTestDeployments'){
          counters=addCrdElem(counters,'name','namespace',clickedId)
        }
        else if  (clickedId == 'skipScalingTestStatefulsets'){
          counters=addCrdElem(counters,'name','namespace',clickedId)
        }
        else{
        if (!counters[clickedId]|| counters[clickedId] <= 1 ) {
                counters[clickedId] = 1;
                var i= counters[clickedId]
                console.log(i)
                console.log(clickedId)
                var field = '<label for="'+clickedId+i+'">'+i+'.</label>'+
                '<pf-text-input id='+clickedId+i+' name="'+clickedId+'"></pf-text-input>';
                $('#'+clickedIdOrg).after(field);
            }
            else{
                var i= counters[clickedId]
                var field = '<label for="'+clickedId+i+'">'+i+'.</label>'+
                '<pf-text-input id='+clickedId+i+' name="'+clickedId+'"></pf-text-input>';
                var target=clickedId+(i-1)+''
                $('#'+target).after(field);
            }
          }
          document.getElementById(clickedId+'remove').style.display = 'block';
          counters[clickedId] = counters[clickedId] + 1; // Increment the counter for the next call
      })
             // Remove last text input
      $('.remove').on('click', function() {
        var clickedIdOrg = $(this).attr('id');
        var clickedId = clickedIdOrg.replace('remove', '');
        var i= counters[clickedId]
        if (i>1){
          var target=clickedId+(i-1)+''
          if (clickedId == 'targetCrdFilters'){
             var target1='targetCrdFiltersnameSuffix'+(i-1)+''
             var target2='targetCrdFiltersscalable'+(i-1)+''
             console.log(target2)
              $('#'+target1).remove();
              $('h7[for="' + target1 + '"]').remove();
              $('#'+target2).remove();
              $('h7[for="' + target2 + '"]').remove();
        }if (clickedId == 'skipScalingTestDeployments'){
             var target1='skipScalingTestDeploymentsname'+(i-1)+''
             var target2='skipScalingTestDeploymentsnamespace'+(i-1)+''
             console.log(target2)
              $('#'+target1).remove();
              $('h7[for="' + target1 + '"]').remove();
              $('#'+target2).remove();
              $('h7[for="' + target2 + '"]').remove();
         }if (clickedId == 'skipScalingTestStatefulsets'){
             var target1='skipScalingTestStatefulsetsname'+(i-1)+''
             var target2='skipScalingTestStatefulsetsnamespace'+(i-1)+''
             console.log(target2)
              $('#'+target1).remove();
              $('h7[for="' + target1 + '"]').remove();
              $('#'+target2).remove();
              $('h7[for="' + target2 + '"]').remove();
         }
         else{
            $('#'+target).remove();
        }
            $('label[for="' + target + '"]').remove();
            counters[clickedId] = counters[clickedId] - 1;
            if (i-1 ==1){
              document.getElementById(clickedIdOrg).style.display = 'none';
            }
      }
        });
    })

    function handleTnfFiles () {
      const fileList = this.files
      if (fileList.length) {
        // We have a file to load
        const fileUploaded = new FileReader()
        fileUploaded.addEventListener('load', e => {
          yamlGlobal = jsyaml.load(fileUploaded.result);
          renderResults()
        })
        fileUploaded.readAsText(fileList[0])
        }
      }

      // render results tab
function renderResults () {
if (typeof yamlGlobal !== 'undefined') {
fillData(yamlGlobal.targetNameSpaces, '#targetNameSpacesadd','targetNameSpaces','name','')
fillData(yamlGlobal.managedDeployments, '#managedDeploymentsadd','managedDeployments','name','')
fillData(yamlGlobal.managedStatefulsets, '#managedStatefulsetsadd','managedStatefulsets','name','')
fillData(yamlGlobal.acceptedKernelTaints, '#acceptedKernelTaintsadd','acceptedKernelTaints','module','')
fillData(yamlGlobal.skipHelmChartList, '#skipHelmChartListadd','skipHelmChartList','name','')

fillData(yamlGlobal.podsUnderTestLabels, '#podsUnderTestLabelsadd','podsUnderTestLabels','','')
fillData(yamlGlobal.servicesignorelist, '#servicesignorelistadd','servicesignorelist','','')
fillData(yamlGlobal.validProtocolNames, '#ValidProtocolNamesadd','ValidProtocolNames','','')
fillData(yamlGlobal.operatorsUnderTestLabels, '#operatorsUnderTestLabelsadd','operatorsUnderTestLabels','','')


fillData(yamlGlobal.skipScalingTestDeployments, '#skipScalingTestDeploymentsadd','skipScalingTestDeployments','name','namespace')
fillData(yamlGlobal.skipScalingTestStatefulsets, '#skipScalingTestStatefulsetsadd','skipScalingTestStatefulsets','name','namespace')
fillData(yamlGlobal.targetCrdFilters, '#targetCrdFiltersadd','targetCrdFilters','nameSuffix','scalable')    
if (yamlGlobal.DebugDaemonSetNamespace){
document.getElementById('DebugDaemonSetNamespace').value = yamlGlobal.DebugDaemonSetNamespace;
}
if (yamlGlobal.DebugDaemonSetNamespace){
  document.getElementById('CollectorAppEndPoint').value = yamlGlobal.collectorAppEndPoint;
}
if(yamlGlobal.executedBy){
  document.getElementById('executedBy').value = yamlGlobal.executedBy;
}
if(yamlGlobal.partnerName){
  document.getElementById('PartnerName').value = yamlGlobal.partnerName;
}
}
}

function fillData(input,element,clickedId,keyname,key2name){
if (!counters[clickedId]){
counters[clickedId] =1
}
for (const key in input) {
var target=clickedId+(counters[clickedId]-1)+''
if (keyname!=''){
  var value = input[key][keyname]
}else{
  var value = input[key]
}
if(key2name==''){
  var pf_txt='<pf-text-input id2='+clickedId+ ' id='+clickedId+counters[clickedId] +' value="'
  +value+'"></pf-text-input>'
}else{
      var i= counters[clickedId]
      var t1=clickedId+keyname+i+''
      var t2=clickedId+key2name+i+''
      console.log(t1)
      var pf_txt =
      '<h7 for="'+t1+'">'+keyname+':</h7><pf-text-input id='+t1+
      ' name="'+clickedId+keyname+'" value="'+ input[key][keyname]+'" ></pf-text-input>' + 
      '<h7 for="'+t2+'">'+key2name+':</h7><pf-text-input id='+t2+
      ' name="'+clickedId+key2name+'" value="'+ input[key][key2name]+'" ></pf-text-input>' ;
      var target=clickedId+key2name+(i-1)+''
}
var field='<label for="'+clickedId+counters[clickedId] +'">'+counters[clickedId] +'.</label>'+pf_txt
 if ( counters[clickedId]==1){
  $(element).after(field);
 }else{
  $('#'+target).after(field);
 }
 counters[clickedId]= counters[clickedId]+1
}
}

 function addCheckedTest(){
  document.getElementById('lifecycle').checked = true
  document.getElementById('manageability').checked = true
  document.getElementById('affiliated-certification').checked = true
  document.getElementById('operator').checked = true
  document.getElementById('access-control').checked = true
  document.getElementById('platform-alteration').checked = true
  document.getElementById('networking').checked = true
  document.getElementById('performance').checked = true
  document.getElementById('observability').checked = true

 } 

function selectScenario(table){
  var idtable = '#'+table+'-table'
  $(idtable).empty()
 // addCheckedTest()
  var field = ""
  const selectScenarioComboBox = document.getElementById('selectScenarioComboBox')
  const selectedScenario = selectScenarioComboBox.options[selectScenarioComboBox.selectedIndex].value
    document.getElementById('selectOpt').setAttribute('hidden','hidden')

  for (const key in classification) {
    if (key.startsWith(table)){
      field=
      '<rh-accordion><rh-accordion-header><div>'+    
      '<input type="checkbox" id ="'+key+ '" value="'+
      key +'" name="selectedOptions" onclick="updateTestCounterTest(\''+table+'\',\''+key+'\',true)" checked> ' +
      '<label for="'+key+ '" id="'+key+'l">'+key+'</label>'+
      '</div></rh-accordion-header>'+
      '<rh-accordion-panel><h4>Description </h4><p>'+
      classification[key][0].description+'</p>'+
      '<h4>Remediation</h4><p>'+
      classification[key][0].remediation+'</p>'+
      '<h4>BestPracticeReference</h4><rh-cta><a href="'+classification[key][0].bestPracticeReference+
      '" target="_blank" rel="noreferrer nofollow">'+
      classification[key][0].bestPracticeReference+'</a></rh-cta></rh-accordion-panel>'+
      '</rh-accordion>'
      $(field).appendTo(idtable);
      document.getElementById(key).addEventListener('click', function(event) {
        // Prevent click propagation from the input element
        event.stopPropagation();
        });
      document.getElementById(key+'l').addEventListener('click', function(event) {
        // Prevent click propagation from the input element
        event.stopPropagation();
        });
    }
  }

  if (selectedScenario != 'all' && selectedScenario != 'none'  ) {
    document.getElementById('selectOpt').removeAttribute('hidden')
  }
  buildTest(selectedScenario,table)
  updateTestCounter(table)
  //updateAllCounter()

}

function updateCheckbox(table){ 
  document.getElementById('selectOpt').setAttribute('hidden','hidden')
  const selectScenarioComboBox = document.getElementById('selectScenarioComboBox')
  const selectedScenario = selectScenarioComboBox.options[selectScenarioComboBox.selectedIndex].value
  if (selectedScenario != 'all' && selectedScenario != 'none'  ) {
    document.getElementById('selectOpt').removeAttribute('hidden')
  }
  buildTest(selectedScenario,table)
  updateTestCounter(table)
  //updateAllCounter()
}

function buildTest(scenarioValue,table){
  const option = document.getElementById('selectOpt')
  const selectedOptionalMandatory = option.options[option.selectedIndex].value
  var eqData = ""
for (const key in classification){
  switch (scenarioValue) {
    case "telco":
      eqData = classification[key][0]["categoryClassification"].Telco;
      break;
    case "nontelco":
      eqData = classification[key][0]["categoryClassification"].NonTelco;
      break;
    case "extended":
      eqData = classification[key][0]["categoryClassification"].Extended;
        break;
    default:
      eqData = classification[key][0]["categoryClassification"].FarEdge;
}
if (key.startsWith(table)){
  var selectedTestItem = document.getElementById(key); 
  if(eqData == selectedOptionalMandatory || scenarioValue == "all"){
    selectedTestItem.checked = true;
      } if(scenarioValue != "all" &&(eqData != selectedOptionalMandatory || scenarioValue == "none")){
        selectedTestItem.checked = false;
      }
  }
}
}

function show(id){
   updateTestCounter(id)
   if (id !="lifecycle"){
    document.getElementById('lifecycle-table').setAttribute('hidden','hidden')

   }
   if (id !="manageability"){
    document.getElementById('manageability-table').setAttribute('hidden','hidden')

   }   
   if (id !="affiliated-certification"){
    document.getElementById('affiliated-certification-table').setAttribute('hidden','hidden')

   }
   if (id !="operator"){
    document.getElementById('operator-table').setAttribute('hidden','hidden')

   }
   if (id !="access-control"){
    document.getElementById('access-control-table').setAttribute('hidden','hidden')

   }
   if (id !="platform-alteration"){
    document.getElementById('platform-alteration-table').setAttribute('hidden','hidden')

   }
   if (id !="networking"){
    document.getElementById('networking-table').setAttribute('hidden','hidden')

   }
   if (id !="performance"){
    document.getElementById('performance-table').setAttribute('hidden','hidden')

   }
   if (id !="observability"){
    document.getElementById('observability-table').setAttribute('hidden','hidden')
   }
   document.getElementById(id+'-table').removeAttribute('hidden')
}

function updateTestCounter(id){
  var CountElement = document.getElementById(id+'-count');
  var matchResult = CountElement.textContent.match(/(.*?:\s)(\d+)/);
  var nonNumericPart = matchResult[1];
  CountElement.textContent = nonNumericPart + 0;

      for (const key in classification) {      
        if (key.startsWith(id)){
          updateTestCounterTest(id,key,false)
          }
    }
  }


  function updateAllCounter(){
   testsList = ['lifecycle','manageability','affiliated-certification','operator'
  ,'access-control','platform-alteration','networking','performance','observability']
  var allCountElement = document.getElementById('all-count');
  var allmatchResult = allCountElement.textContent.match(/(.*?:\s)(\d+)/);
  var allnonNumericPart = allmatchResult[1];
  var count = 0
    for (var i = 0; i < testsList.length; i++) {
      var test =  testsList[i]
      var CountElement = document.getElementById(test+'-count');
      var matchResult = CountElement.textContent.match(/(.*?:\s)(\d+)/);
      var newCount = parseInt(matchResult[2]); 
      count +=newCount
    }
    allCountElement.textContent = allnonNumericPart + count;
    }
  
  function updateTestCounterTest(id,key,comfromsmallCombox){
    var CountElement = document.getElementById(id+'-count');
    var matchResult = CountElement.textContent.match(/(.*?:\s)(\d+)/);
    var newCount = parseInt(matchResult[2]); 
    var nonNumericPart = matchResult[1];
    var checkedkey = document.getElementById(key);
      if (checkedkey.checked) {
       newCount = newCount + 1; 
      }else{ 
  if (comfromsmallCombox){
    newCount = newCount - 1; 

  }
}
  if (newCount<0){
    newCount=0
  }
  CountElement.textContent = nonNumericPart + newCount;
  updateAllCounter()
  }

function show(id){
   updateTestCounter(id)
   if (id !="lifecycle"){
    document.getElementById('lifecycle-table').setAttribute('hidden','hidden')

   }
   if (id !="manageability"){
    document.getElementById('manageability-table').setAttribute('hidden','hidden')

   }   
   if (id !="affiliated-certification"){
    document.getElementById('affiliated-certification-table').setAttribute('hidden','hidden')

   }
   if (id !="operator"){
    document.getElementById('operator-table').setAttribute('hidden','hidden')

   }
   if (id !="access-control"){
    document.getElementById('access-control-table').setAttribute('hidden','hidden')

   }
   if (id !="platform-alteration"){
    document.getElementById('platform-alteration-table').setAttribute('hidden','hidden')

   }
   if (id !="networking"){
    document.getElementById('networking-table').setAttribute('hidden','hidden')

   }
   if (id !="performance"){
    document.getElementById('performance-table').setAttribute('hidden','hidden')

   }
   if (id !="observability"){
    document.getElementById('observability-table').setAttribute('hidden','hidden')
   }
   document.getElementById(id+'-table').removeAttribute('hidden')
}

function updateTestCounter(id){
  var CountElement = document.getElementById(id+'-count');
  var matchResult = CountElement.textContent.match(/(.*?:\s)(\d+)/);
  var nonNumericPart = matchResult[1];
  CountElement.textContent = nonNumericPart + 0;

      for (const key in classification) {      
        if (key.startsWith(id)){
          updateTestCounterTest(id,key)
          }
    }
    //updateAllCounter()
  }


  function updateAllCounter(){
   testsList = ['lifecycle','manageability','affiliated-certification','operator'
  ,'access-control','platform-alteration','networking','performance','observability']
  var allCountElement = document.getElementById('all-count');
  var allmatchResult = allCountElement.textContent.match(/(.*?:\s)(\d+)/);
  var allnonNumericPart = allmatchResult[1];
  var count = 0
    for (var i = 0; i < testsList.length; i++) {
      var test =  testsList[i]
      var CountElement = document.getElementById(test+'-count');
      var matchResult = CountElement.textContent.match(/(.*?:\s)(\d+)/);
      var newCount = parseInt(matchResult[2]); 
      count +=newCount
    }
    allCountElement.textContent = allnonNumericPart + count;
    }
  
  function updateTestCounterTest(id,key,comfromsmallCombox){
    var CountElement = document.getElementById(id+'-count');
    var matchResult = CountElement.textContent.match(/(.*?:\s)(\d+)/);
    var newCount = parseInt(matchResult[2]); 
    var nonNumericPart = matchResult[1];
    var checkedkey = document.getElementById(key);
      if (checkedkey.checked) {
       newCount = newCount + 1; 
      }else{ 
  if (comfromsmallCombox){
    newCount = newCount - 1; 

  }
}
  if (newCount<0){
    newCount=0
  }
  CountElement.textContent = nonNumericPart + newCount;
  updateAllCounter()
  }

// Add an event listener to the checkbox
function performToggle (triggerId) {
  var checkbox = document.getElementById(triggerId);
  for (const key in classification) {  
    if (key.startsWith(triggerId)){
    if (checkbox.checked) {
      var rowToRemove = document.getElementById(key);
      if (rowToRemove) {
      rowToRemove.checked = true;
      }


    } else {
      var rowToRemove = document.getElementById(key); 
      // Check if the row exists
      if (rowToRemove) {
          // Remove the row from the table
          rowToRemove.checked = false;
        }
    }
  }
}
updateTestCounter(triggerId)
//updateAllCounter()
}